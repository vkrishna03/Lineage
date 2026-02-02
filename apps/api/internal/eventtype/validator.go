package eventtype

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

// validIntents are the allowed event intent values
var validIntents = map[string]bool{
	"exploration": true,
	"suggestion":  true,
	"assertion":   true,
	"decision":    true,
	"execution":   true,
}

// ValidatePayloadSchema validates that a payload schema is a valid JSON schema
func ValidatePayloadSchema(schema json.RawMessage) error {
	if len(schema) == 0 {
		return nil // No schema is valid (optional field)
	}

	// First check it's valid JSON
	var schemaData interface{}
	if err := json.Unmarshal(schema, &schemaData); err != nil {
		return fmt.Errorf("payload_schema is not valid JSON: %w", err)
	}

	// Try to compile as JSON schema
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource("schema.json", bytes.NewReader(schema)); err != nil {
		return fmt.Errorf("payload_schema is not a valid JSON schema: %w", err)
	}

	if _, err := compiler.Compile("schema.json"); err != nil {
		return fmt.Errorf("payload_schema failed to compile: %w", err)
	}

	return nil
}

// ValidateAllowedIntents validates that all intents in the list are valid
func ValidateAllowedIntents(intents []string) error {
	if len(intents) == 0 {
		return nil // Empty list is valid (all intents allowed)
	}

	for _, intent := range intents {
		if !validIntents[intent] {
			return fmt.Errorf("invalid intent: %s (valid: exploration, suggestion, assertion, decision, execution)", intent)
		}
	}

	return nil
}
