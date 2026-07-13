package tools

import (
	"encoding/json"
)

// ToolResult encapsula a resposta detalhada para o LLM e a resposta resumida para o usuário.
type ToolResult struct {
	FullResult         string // O conteúdo completo (para o LLM)
	Summary           string // O sumário curto (ex: "2 arquivos lidos", "Testes passados")
	RequiresFullOutput bool   // Se true, a TUI deve exibir o FullResult mesmo fora do modo verboso (ex: erro de teste)
}

// Tool define a interface para todas as ferramentas do agente.
type Tool interface {
	Name() string
	Execute(args map[string]interface{}) (ToolResult, error)
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
