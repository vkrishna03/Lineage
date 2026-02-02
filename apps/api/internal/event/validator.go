package event

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

// Validator validates event payloads against JSON schemas
type Validator struct{}

// NewValidator creates a new Validator
func NewValidator() *Validator {
	return &Validator{}
}

// ValidatePayload validates a payload against a JSON schema
func (v *Validator) ValidatePayload(payload json.RawMessage, schema []byte) error {
	if len(schema) == 0 {
		return nil // No schema = no validation required
	}

	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource("schema.json", bytes.NewReader(schema)); err != nil {
		return fmt.Errorf("invalid schema: %w", err)
	}

	sch, err := compiler.Compile("schema.json")
	if err != nil {
		return fmt.Errorf("failed to compile schema: %w", err)
	}

	var data interface{}
	if err := json.Unmarshal(payload, &data); err != nil {
		return fmt.Errorf("invalid JSON payload: %w", err)
	}

	if err := sch.Validate(data); err != nil {
		return fmt.Errorf("payload validation failed: %w", err)
	}

	return nil
}
