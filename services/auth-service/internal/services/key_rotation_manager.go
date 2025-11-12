package services

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
	"github.com/v-egorov/service-boilerplate/services/auth-service/internal/utils"
)

// RotationConfig holds configuration for key rotation
type RotationConfig struct {
	Enabled        bool          `yaml:"enabled"`
	Type           string        `yaml:"type"` // "time", "usage", "manual"
	IntervalDays   int           `yaml:"interval_days"`
	MaxTokens      int           `yaml:"max_tokens"`
	OverlapMinutes int           `yaml:"overlap_minutes"`
	CheckInterval  time.Duration `yaml:"check_interval"` // How often to check for rotation
}

// KeyRotationManager handles automatic JWT key rotation
type KeyRotationManager struct {
	jwtUtils *utils.JWTUtils
	db       *pgxpool.Pool
	config   RotationConfig
	logger   *logrus.Logger
	stopCh   chan struct{}
}

// NewKeyRotationManager creates a new key rotation manager
func NewKeyRotationManager(jwtUtils *utils.JWTUtils, db *pgxpool.Pool, config RotationConfig, logger *logrus.Logger) *KeyRotationManager {
	return &KeyRotationManager{
		jwtUtils: jwtUtils,
		db:       db,
		config:   config,
		logger:   logger,
		stopCh:   make(chan struct{}),
	}
}

// Start begins the automatic key rotation process
func (krm *KeyRotationManager) Start(ctx context.Context) {
	if !krm.config.Enabled {
		krm.logger.Info("JWT key rotation is disabled")
		return
	}

	krm.logger.WithFields(logrus.Fields{
		"type":           krm.config.Type,
		"interval_days":  krm.config.IntervalDays,
		"check_interval": krm.config.CheckInterval,
	}).Info("Starting JWT key rotation manager")

	go krm.rotationLoop(ctx)
}

// Stop stops the key rotation manager
func (krm *KeyRotationManager) Stop() {
	close(krm.stopCh)
}

// rotationLoop runs the main rotation check loop
func (krm *KeyRotationManager) rotationLoop(ctx context.Context) {
	ticker := time.NewTicker(krm.config.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			krm.logger.Info("Key rotation manager stopping due to context cancellation")
			return
		case <-krm.stopCh:
			krm.logger.Info("Key rotation manager stopping")
			return
		case <-ticker.C:
			if err := krm.checkAndRotate(ctx); err != nil {
				krm.logger.WithError(err).Error("Failed to check and rotate keys")
			}
		}
	}
}

// checkAndRotate checks if rotation is needed and performs it
func (krm *KeyRotationManager) checkAndRotate(ctx context.Context) error {
	shouldRotate, reason, err := krm.shouldRotate(ctx)
	if err != nil {
		return fmt.Errorf("failed to check rotation status: %w", err)
	}

	if shouldRotate {
		krm.logger.WithField("reason", reason).Info("Initiating automatic key rotation")
		return krm.performRotation(ctx, reason)
	}

	return nil
}

// shouldRotate determines if key rotation is needed
func (krm *KeyRotationManager) shouldRotate(ctx context.Context) (bool, string, error) {
	switch krm.config.Type {
	case "time":
		return krm.shouldRotateByTime(ctx)
	case "usage":
		return krm.shouldRotateByUsage(ctx)
	default:
		return false, "", nil
	}
}

// shouldRotateByTime checks if rotation is needed based on time interval
func (krm *KeyRotationManager) shouldRotateByTime(ctx context.Context) (bool, string, error) {
	query := `
		SELECT rotated_at
		FROM auth_service.jwt_keys
		WHERE is_active = true
		ORDER BY created_at DESC
		LIMIT 1
	`

	var rotatedAt *time.Time
	err := krm.db.QueryRow(ctx, query).Scan(&rotatedAt)
	if err != nil {
		return false, "", fmt.Errorf("failed to get current key rotation time: %w", err)
	}

	if rotatedAt == nil {
		return true, "no_rotation_time_set", nil
	}

	timeSinceRotation := time.Since(*rotatedAt)
	rotationInterval := time.Duration(krm.config.IntervalDays) * 24 * time.Hour

	if timeSinceRotation >= rotationInterval {
		return true, fmt.Sprintf("time_interval_exceeded_%v", timeSinceRotation.Round(time.Hour)), nil
	}

	return false, "", nil
}

// shouldRotateByUsage checks if rotation is needed based on token usage
func (krm *KeyRotationManager) shouldRotateByUsage(ctx context.Context) (bool, string, error) {
	// This would require tracking token issuance count per key
	// For now, return false as this is not implemented yet
	return false, "", nil
}

// performRotation performs the actual key rotation
func (krm *KeyRotationManager) performRotation(ctx context.Context, reason string) error {
	// Update the rotation reason in the current key before rotating
	currentKeyID := krm.jwtUtils.GetKeyID()
	if currentKeyID != "" {
		updateQuery := `
			UPDATE auth_service.jwt_keys
			SET rotation_reason = $1
			WHERE key_id = $2 AND is_active = true
		`
		_, err := krm.db.Exec(ctx, updateQuery, reason, currentKeyID)
		if err != nil {
			krm.logger.WithError(err).Warn("Failed to update rotation reason for current key")
		}
	}

	// Perform the rotation
	if err := krm.jwtUtils.RotateKeysWithReason(ctx, reason); err != nil {
		return fmt.Errorf("failed to rotate keys: %w", err)
	}

	// Update the rotation config with last rotation time
	updateConfigQuery := `
		UPDATE auth_service.key_rotation_config
		SET last_rotation_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE enabled = true
	`
	_, err := krm.db.Exec(ctx, updateConfigQuery)
	if err != nil {
		krm.logger.WithError(err).Warn("Failed to update rotation config timestamp")
	}

	krm.logger.WithFields(logrus.Fields{
		"reason":  reason,
		"new_key": krm.jwtUtils.GetKeyID(),
	}).Info("JWT key rotation completed successfully")

	return nil
}

// GetRotationStatus returns the current rotation status
func (krm *KeyRotationManager) GetRotationStatus(ctx context.Context) (map[string]interface{}, error) {
	query := `
		SELECT
			j.key_id,
			j.rotated_at,
			j.rotation_reason,
			EXTRACT(EPOCH FROM (CURRENT_TIMESTAMP - j.rotated_at)) / 86400 as days_since_rotation,
			c.rotation_type,
			c.interval_days,
			c.enabled,
			c.last_rotation_at
		FROM auth_service.jwt_keys j
		CROSS JOIN auth_service.key_rotation_config c
		WHERE j.is_active = true
		ORDER BY j.created_at DESC
		LIMIT 1
	`

	var keyID string
	var rotatedAt *time.Time
	var rotationReason *string
	var daysSinceRotation *float64
	var rotationType string
	var intervalDays int
	var enabled bool
	var lastRotationAt *time.Time

	err := krm.db.QueryRow(ctx, query).Scan(
		&keyID, &rotatedAt, &rotationReason, &daysSinceRotation,
		&rotationType, &intervalDays, &enabled, &lastRotationAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get rotation status: %w", err)
	}

	status := map[string]interface{}{
		"current_key_id":      keyID,
		"rotated_at":          rotatedAt,
		"rotation_reason":     rotationReason,
		"days_since_rotation": daysSinceRotation,
		"rotation_type":       rotationType,
		"interval_days":       intervalDays,
		"enabled":             enabled,
		"last_rotation_at":    lastRotationAt,
		"next_rotation_due":   krm.calculateNextRotation(rotatedAt, intervalDays),
	}

	return status, nil
}

// calculateNextRotation calculates when the next rotation is due
func (krm *KeyRotationManager) calculateNextRotation(rotatedAt *time.Time, intervalDays int) *time.Time {
	if rotatedAt == nil {
		return nil
	}

	nextRotation := rotatedAt.Add(time.Duration(intervalDays) * 24 * time.Hour)
	return &nextRotation
}
