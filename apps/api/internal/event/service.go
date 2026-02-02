package event

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/cyberphone/json-canonicalization/go/src/webpki.org/jsoncanonicalizer"
	"github.com/google/uuid"
)

// Input represents the data sent by producers to create an event
type Input struct {
	ScopeID          uuid.UUID       `json:"scope_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	ActorID          uuid.UUID       `json:"actor_id" example:"550e8400-e29b-41d4-a716-446655440001"`
	EventTypeID      uuid.UUID       `json:"event_type_id" example:"550e8400-e29b-41d4-a716-446655440002"`
	Intent           string          `json:"intent" example:"decision" enums:"exploration,suggestion,assertion,decision,execution"`
	Reason           *string         `json:"reason,omitempty" example:"User approved the recommendation"`
	CorrectionType   *string         `json:"correction_type,omitempty" example:"supersede" enums:"supersede,amend,retract"`
	CorrectsEventID  *uuid.UUID      `json:"corrects_event_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440003"`
	ObservedAt       *time.Time      `json:"observed_at,omitempty" example:"2024-01-15T10:30:00Z"`
	DecidedAt        *time.Time      `json:"decided_at,omitempty" example:"2024-01-15T10:35:00Z"`
	Payload          json.RawMessage `json:"payload" swaggertype:"object"`
	ParentEventIDs   []uuid.UUID     `json:"parent_event_ids,omitempty"`

	// Inline confidence score (optional, 0.0-1.0)
	Confidence *float64 `json:"confidence,omitempty" example:"0.85"`

	// Inline artifact linking (optional)
	InputArtifactIDs  []uuid.UUID `json:"input_artifact_ids,omitempty"`
	OutputArtifactIDs []uuid.UUID `json:"output_artifact_ids,omitempty"`
}

// Hashable contains fields used for hash computation per RFC 8785
// Excluded: id, ingested_at, prev_event_hash, event_hash (assigned at write time)
type Hashable struct {
	ScopeID         uuid.UUID       `json:"scope_id"`
	ActorID         uuid.UUID       `json:"actor_id"`
	EventTypeID     uuid.UUID       `json:"event_type_id"`
	Intent          string          `json:"intent"`
	ScopeSequence   int64           `json:"scope_sequence"`
	CorrectionType  *string         `json:"correction_type,omitempty"`
	CorrectsEventID *uuid.UUID      `json:"corrects_event_id,omitempty"`
	ObservedAt      *time.Time      `json:"observed_at,omitempty"`
	DecidedAt       *time.Time      `json:"decided_at,omitempty"`
	Reason          *string         `json:"reason,omitempty"`
	Payload         json.RawMessage `json:"payload"`
}

// ComputeHash computes SHA-256 hash of canonical JSON (RFC 8785)
func ComputeHash(event Hashable) (string, error) {
	jsonBytes, err := json.Marshal(event)
	if err != nil {
		return "", err
	}

	canonical, err := jsoncanonicalizer.Transform(jsonBytes)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(canonical)
	return hex.EncodeToString(hash[:]), nil
}

// ScoreCategory represents the categorical bucket for a score value
type ScoreCategory string

const (
	ScoreCategoryVeryLow  ScoreCategory = "very_low"
	ScoreCategoryLow      ScoreCategory = "low"
	ScoreCategoryModerate ScoreCategory = "moderate"
	ScoreCategoryHigh     ScoreCategory = "high"
	ScoreCategoryVeryHigh ScoreCategory = "very_high"
)

// DeriveScoreCategory derives category from numeric value (0.0-1.0)
func DeriveScoreCategory(value float64) ScoreCategory {
	switch {
	case value < 0.20:
		return ScoreCategoryVeryLow
	case value < 0.40:
		return ScoreCategoryLow
	case value < 0.60:
		return ScoreCategoryModerate
	case value < 0.80:
		return ScoreCategoryHigh
	default:
		return ScoreCategoryVeryHigh
	}
}

// ValidIntents are the allowed event intent values
var ValidIntents = map[string]bool{
	"exploration": true,
	"suggestion":  true,
	"assertion":   true,
	"decision":    true,
	"execution":   true,
}

// ValidCorrectionTypes are the allowed correction type values
var ValidCorrectionTypes = map[string]bool{
	"supersede": true,
	"amend":     true,
	"retract":   true,
}

// ValidateIntent checks if the intent value is valid
func ValidateIntent(intent string) bool {
	return ValidIntents[intent]
}

// ValidateCorrectionType checks if the correction type value is valid
func ValidateCorrectionType(ct string) bool {
	return ValidCorrectionTypes[ct]
}
