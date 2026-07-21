package llm

import (
	"encoding/json"
	"testing"
)

func TestToolCallFunction_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedName string
		expectedArgs map[string]interface{}
		wantErr      bool
	}{
		{
			name:         "Standard JSON arguments",
			input:        `{"name": "read_file", "arguments": {"path": "test.txt"}}`,
			expectedName: "read_file",
			expectedArgs: map[string]interface{}{"path": "test.txt"},
			wantErr:      false,
		},
		{
			name: "Double-encoded JSON string arguments",
			// This represents a JSON object where the 'arguments' field is a string containing JSON.
			// In raw JSON: {"name": "read_file", "arguments": "{\"path\": \"test.txt\"}"}
			input:        `{"name": "read_file", "arguments": "{\"path\": \"test.txt\"}"}`,
			expectedName: "read_file",
			expectedArgs: map[string]interface{}{"path": "test.txt"},
			wantErr:      false,
		},
		{
			name:         "Malformed JSON",
			input:        `{"name": "read_file", "arguments": {"path": "test.txt"`, // Missing closing brace
			expectedName: "",
			expectedArgs: nil,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tf ToolCallFunction
			err := json.Unmarshal([]byte(tt.input), &tf)

			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if tf.Name != tt.expectedName {
					t.Errorf("Expected name %s, got %s", tt.expectedName, tf.Name)
				}

				var actualArgs map[string]interface{}
				if err := json.Unmarshal(tf.Arguments, &actualArgs); err != nil {
					t.Fatalf("Failed to unmarshal resulting arguments: %v", err)
				}

				if len(actualArgs) != len(tt.expectedArgs) {
					t.Errorf("Expected %d args, got %d", len(tt.expectedArgs), len(actualArgs))
				}

				for k, v := range tt.expectedArgs {
					if actualArgs[k] != v {
						t.Errorf("Expected arg %s=%v, got %v", k, v, actualArgs[k])
					}
				}
			}
		})
	}
}
