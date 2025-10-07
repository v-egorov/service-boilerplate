package utils

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type JWTClaims struct {
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	Roles     []string  `json:"roles"`
	TokenType string    `json:"token_type"`
	jwt.RegisteredClaims
}

// JWTKey represents a stored JWT key in the database
type JWTKey struct {
	ID             uuid.UUID       `db:"id"`
	KeyID          string          `db:"key_id"`
	PrivateKeyPEM  string          `db:"private_key_pem"`
	PublicKeyPEM   string          `db:"public_key_pem"`
	Algorithm      string          `db:"algorithm"`
	IsActive       bool            `db:"is_active"`
	CreatedAt      time.Time       `db:"created_at"`
	ExpiresAt      *time.Time      `db:"expires_at"`
	RotationReason *string         `db:"rotation_reason"`
	RotatedAt      *time.Time      `db:"rotated_at"`
	Metadata       json.RawMessage `db:"metadata"`
}

type JWTUtils struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	keyID      string
	db         *pgxpool.Pool
}

func NewJWTUtils(db *pgxpool.Pool) (*JWTUtils, error) {
	utils := &JWTUtils{db: db}

	// Try to load existing active key from database
	existingKey, err := utils.loadActiveKey(context.Background())
	if err == nil && existingKey != nil {
		// Successfully loaded existing key
		privateKey, err := utils.parsePrivateKeyPEM(existingKey.PrivateKeyPEM)
		if err != nil {
			return nil, fmt.Errorf("failed to parse existing private key: %w", err)
		}

		utils.privateKey = privateKey
		utils.publicKey = &privateKey.PublicKey
		utils.keyID = existingKey.KeyID
		return utils, nil
	}

	// No existing key found, generate new one
	return utils.generateAndStoreNewKey(context.Background())
}

func (j *JWTUtils) generateAndStoreNewKeyWithReason(ctx context.Context, reason string) (*JWTUtils, error) {
	// Generate RSA key pair for JWT signing
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key: %w", err)
	}

	// Generate unique key ID
	keyID := fmt.Sprintf("jwt-key-%s", uuid.New().String()[:8])

	// Convert keys to PEM format
	privateKeyPEM, err := j.privateKeyToPEM(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encode private key: %w", err)
	}

	publicKeyPEM, err := j.publicKeyToPEM(&privateKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encode public key: %w", err)
	}

	// Store in database
	now := time.Now()
	jwtKey := &JWTKey{
		KeyID:          keyID,
		PrivateKeyPEM:  string(privateKeyPEM),
		PublicKeyPEM:   string(publicKeyPEM),
		Algorithm:      "RS256",
		IsActive:       true,
		RotationReason: &reason,
		RotatedAt:      &now,
		Metadata:       json.RawMessage(`{"key_size": 2048}`),
	}

	if err := j.storeKey(ctx, jwtKey); err != nil {
		return nil, fmt.Errorf("failed to store JWT key: %w", err)
	}

	return &JWTUtils{
		privateKey: privateKey,
		publicKey:  &privateKey.PublicKey,
		keyID:      keyID,
		db:         j.db,
	}, nil
}

func (j *JWTUtils) generateAndStoreNewKey(ctx context.Context) (*JWTUtils, error) {
	return j.generateAndStoreNewKeyWithReason(ctx, "manual")
}

func (j *JWTUtils) GenerateAccessToken(userID uuid.UUID, email string, roles []string, expiration time.Duration) (string, error) {
	now := time.Now()
	claims := JWTClaims{
		UserID:    userID,
		Email:     email,
		Roles:     roles,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "auth-service",
			Subject:   userID.String(),
			Audience:  jwt.ClaimStrings{"api"},
			ExpiresAt: jwt.NewNumericDate(now.Add(expiration)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(j.privateKey)
}

func (j *JWTUtils) GenerateRefreshToken(userID uuid.UUID, expiration time.Duration) (string, error) {
	now := time.Now()
	claims := JWTClaims{
		UserID:    userID,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "auth-service",
			Subject:   userID.String(),
			Audience:  jwt.ClaimStrings{"api"},
			ExpiresAt: jwt.NewNumericDate(now.Add(expiration)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(j.privateKey)
}

func (j *JWTUtils) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.publicKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

func (j *JWTUtils) GetPublicKeyPEM() ([]byte, error) {
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(j.publicKey)
	if err != nil {
		return nil, err
	}

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	return publicKeyPEM, nil
}

func (j *JWTUtils) GetPrivateKeyPEM() ([]byte, error) {
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(j.privateKey)

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	return privateKeyPEM, nil
}

func (j *JWTUtils) GetPublicKey() *rsa.PublicKey {
	return j.publicKey
}

func (j *JWTUtils) GetKeyID() string {
	return j.keyID
}

// RotateKeys generates a new RSA key pair and replaces the current active key
func (j *JWTUtils) RotateKeys(ctx context.Context) error {
	return j.RotateKeysWithReason(ctx, "manual")
}

// RotateKeysWithReason generates a new RSA key pair and replaces the current active key with a specific reason
func (j *JWTUtils) RotateKeysWithReason(ctx context.Context, reason string) error {
	newUtils, err := j.generateAndStoreNewKeyWithReason(ctx, reason)
	if err != nil {
		return fmt.Errorf("failed to generate and store new key: %w", err)
	}

	// Update the current instance with the new key
	j.privateKey = newUtils.privateKey
	j.publicKey = newUtils.publicKey
	j.keyID = newUtils.keyID

	return nil
}

// loadActiveKey loads the active JWT key from the database
func (j *JWTUtils) loadActiveKey(ctx context.Context) (*JWTKey, error) {
	query := `
		SELECT id, key_id, private_key_pem, public_key_pem, algorithm, is_active, created_at, expires_at, rotation_reason, rotated_at, metadata
		FROM auth_service.jwt_keys
		WHERE is_active = true
		ORDER BY created_at DESC
		LIMIT 1
	`

	var key JWTKey
	err := j.db.QueryRow(ctx, query).Scan(
		&key.ID, &key.KeyID, &key.PrivateKeyPEM, &key.PublicKeyPEM,
		&key.Algorithm, &key.IsActive, &key.CreatedAt, &key.ExpiresAt, &key.RotationReason, &key.RotatedAt, &key.Metadata,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // No active key found
		}
		return nil, fmt.Errorf("failed to load active JWT key: %w", err)
	}

	return &key, nil
}

// storeKey stores a JWT key in the database
func (j *JWTUtils) storeKey(ctx context.Context, key *JWTKey) error {
	// First, deactivate any existing active keys
	_, err := j.db.Exec(ctx, "UPDATE auth_service.jwt_keys SET is_active = false WHERE is_active = true")
	if err != nil {
		return fmt.Errorf("failed to deactivate existing keys: %w", err)
	}

	// Insert the new key
	query := `
		INSERT INTO auth_service.jwt_keys (key_id, private_key_pem, public_key_pem, algorithm, is_active, rotation_reason, rotated_at, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err = j.db.Exec(ctx, query,
		key.KeyID, key.PrivateKeyPEM, key.PublicKeyPEM, key.Algorithm, key.IsActive, key.RotationReason, key.RotatedAt, key.Metadata)
	if err != nil {
		return fmt.Errorf("failed to insert JWT key: %w", err)
	}

	return nil
}

// parsePrivateKeyPEM parses a PEM-encoded private key
func (j *JWTUtils) parsePrivateKeyPEM(pemData string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	if block.Type != "RSA PRIVATE KEY" {
		return nil, fmt.Errorf("unexpected PEM block type: %s", block.Type)
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return privateKey, nil
}

// privateKeyToPEM converts an RSA private key to PEM format
func (j *JWTUtils) privateKeyToPEM(privateKey *rsa.PrivateKey) ([]byte, error) {
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	return pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}), nil
}

// publicKeyToPEM converts an RSA public key to PEM format
func (j *JWTUtils) publicKeyToPEM(publicKey *rsa.PublicKey) ([]byte, error) {
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return nil, err
	}

	return pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicKeyBytes,
	}), nil
}
