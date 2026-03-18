package models

import (
	"encoding/json"
	"time"
)

// RelationshipType represents a relationship type in the CTI pattern
type RelationshipType struct {
	ObjectID         int64           `json:"object_id" db:"object_id"`
	TypeKey          string          `json:"type_key" db:"type_key"`
	RelationshipName string          `json:"relationship_name" db:"relationship_name"`
	ReverseTypeKey   *string         `json:"reverse_type_key,omitempty" db:"reverse_type_key"`
	Cardinality      string          `json:"cardinality" db:"cardinality"`
	Required         bool            `json:"required" db:"required"`
	MinCount         int             `json:"min_count" db:"min_count"`
	MaxCount         int             `json:"max_count" db:"max_count"`
	ValidationRules  json.RawMessage `json:"validation_rules" db:"validation_rules"`
	CreatedBy        *string         `json:"created_by,omitempty" db:"created_by"`
	UpdatedBy        *string         `json:"updated_by,omitempty" db:"updated_by"`
	CreatedAt        time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at" db:"updated_at"`
}

// TableName returns the table name for RelationshipType model
func (rt *RelationshipType) TableName() string {
	return "objects_relationship_types"
}

// Valid cardinality values
const (
	CardinalityOneToOne   = "one_to_one"
	CardinalityOneToMany  = "one_to_many"
	CardinalityManyToOne  = "many_to_one"
	CardinalityManyToMany = "many_to_many"
)

// ValidCardinalityValues returns all valid cardinality values
func ValidCardinalityValues() []string {
	return []string{
		CardinalityOneToOne,
		CardinalityOneToMany,
		CardinalityManyToOne,
		CardinalityManyToMany,
	}
}

// IsValidCardinality checks if the given cardinality is valid
func IsValidCardinality(cardinality string) bool {
	for _, valid := range ValidCardinalityValues() {
		if cardinality == valid {
			return true
		}
	}
	return false
}

// GetValidationRulesMap converts JSONB validation_rules to map[string]interface{}
func (rt *RelationshipType) GetValidationRulesMap() map[string]interface{} {
	if len(rt.ValidationRules) == 0 {
		return make(map[string]interface{})
	}

	var rules map[string]interface{}
	if err := json.Unmarshal(rt.ValidationRules, &rules); err != nil {
		return make(map[string]interface{})
	}
	return rules
}

// SetValidationRulesMap sets validation_rules from map[string]interface{} to JSONB format
func (rt *RelationshipType) SetValidationRulesMap(rules map[string]interface{}) error {
	if rules == nil {
		rt.ValidationRules = json.RawMessage("{}")
		return nil
	}

	data, err := json.Marshal(rules)
	if err != nil {
		return err
	}
	rt.ValidationRules = data
	return nil
}

// CreateRelationshipTypeRequest represents creation input
type CreateRelationshipTypeRequest struct {
	TypeKey          string                 `json:"type_key" binding:"required" validate:"required,min=1,max=100"`
	RelationshipName string                 `json:"relationship_name" validate:"max=255"`
	ReverseTypeKey   *string                `json:"reverse_type_key,omitempty" validate:"omitempty,min=1,max=100"`
	Cardinality      string                 `json:"cardinality" binding:"required" validate:"required"`
	Required         bool                   `json:"required"`
	MinCount         int                    `json:"min_count"`
	MaxCount         int                    `json:"max_count"`
	ValidationRules  map[string]interface{} `json:"validation_rules"`
	CreatedBy        string                 `json:"-"`
	UpdatedBy        string                 `json:"-"`
}

// UpdateRelationshipTypeRequest represents update input
type UpdateRelationshipTypeRequest struct {
	RelationshipName *string                 `json:"relationship_name,omitempty" validate:"omitempty,max=255"`
	ReverseTypeKey   *string                 `json:"reverse_type_key,omitempty" validate:"omitempty,min=1,max=100"`
	Cardinality      *string                 `json:"cardinality,omitempty"`
	Required         *bool                   `json:"required,omitempty"`
	MinCount         *int                    `json:"min_count,omitempty"`
	MaxCount         *int                    `json:"max_count,omitempty"`
	ValidationRules  *map[string]interface{} `json:"validation_rules,omitempty"`
	UpdatedBy        string                  `json:"-"`
}

// RelationshipTypeFilter for listing
type RelationshipTypeFilter struct {
	Cardinality string `json:"cardinality,omitempty" form:"cardinality"`
	Required    *bool  `json:"required,omitempty" form:"required"`
	Page        int    `json:"page,omitempty" form:"page"`
	PageSize    int    `json:"page_size,omitempty" form:"page_size"`
	SortBy      string `json:"sort_by,omitempty" form:"sort_by"`
	SortOrder   string `json:"sort_order,omitempty" form:"sort_order"`
}

// RelationshipTypeResponse represents the response payload
type RelationshipTypeResponse struct {
	ObjectID         int64                  `json:"object_id"`
	TypeKey          string                 `json:"type_key"`
	RelationshipName string                 `json:"relationship_name"`
	ReverseTypeKey   *string                `json:"reverse_type_key,omitempty"`
	Cardinality      string                 `json:"cardinality"`
	Required         bool                   `json:"required"`
	MinCount         int                    `json:"min_count"`
	MaxCount         int                    `json:"max_count"`
	ValidationRules  map[string]interface{} `json:"validation_rules"`
	CreatedAt        string                 `json:"created_at"`
	UpdatedAt        string                 `json:"updated_at"`
	CreatedBy        string                 `json:"created_by,omitempty"`
	UpdatedBy        string                 `json:"updated_by,omitempty"`
}

// ToResponse converts RelationshipType to RelationshipTypeResponse
func (rt *RelationshipType) ToResponse() *RelationshipTypeResponse {
	var rules map[string]interface{}
	if len(rt.ValidationRules) > 0 {
		if err := json.Unmarshal(rt.ValidationRules, &rules); err != nil {
			rules = make(map[string]interface{})
		}
	} else {
		rules = make(map[string]interface{})
	}

	var createdBy, updatedBy string
	if rt.CreatedBy != nil {
		createdBy = *rt.CreatedBy
	}
	if rt.UpdatedBy != nil {
		updatedBy = *rt.UpdatedBy
	}

	return &RelationshipTypeResponse{
		ObjectID:         rt.ObjectID,
		TypeKey:          rt.TypeKey,
		RelationshipName: rt.RelationshipName,
		ReverseTypeKey:   rt.ReverseTypeKey,
		Cardinality:      rt.Cardinality,
		Required:         rt.Required,
		MinCount:         rt.MinCount,
		MaxCount:         rt.MaxCount,
		ValidationRules:  rules,
		CreatedAt:        rt.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        rt.UpdatedAt.Format(time.RFC3339),
		CreatedBy:        createdBy,
		UpdatedBy:        updatedBy,
	}
}

// RelationshipTypeListResponse represents paginated response
type RelationshipTypeListResponse struct {
	Data       []RelationshipTypeResponse `json:"data"`
	Pagination PaginationResponse         `json:"pagination"`
}
