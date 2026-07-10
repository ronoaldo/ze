package tools

import (
	"encoding/json"
)

// Tool defines the interface for all agent tools.
type Tool interface {
	Name() string
	Execute(args map[string]interface{}) (string, error)
	JSONSchema() map[string]interface{}
}

type FileReadArgs struct {
	Path string `json:"path"`
}

type FileWriteArgs struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type RemoveFileArgs struct {
	Path string `json:"path"`
}

type GoDocArgs struct {
	Package string `json:"package"`
}

type Edit struct {
	OldString  string `json:"oldString"`
	NewString  string `json:"newString"`
	ReplaceAll bool   `json:"replaceAll"`
}

type EditArgs struct {
	Path  string `json:"path"`
	Edits []Edit `json:"edits"`
}

// mapToStruct is a helper to decode map into a struct using JSON tags.
func mapToStruct(m map[string]interface{}, s interface{}) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, s)
}
