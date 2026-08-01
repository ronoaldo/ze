package tools

import (
	"encoding/json"
)

// ToolResult encapsula a resposta detalhada para o LLM e a resposta resumida para o usuário.
type ToolResult struct {
	FullResult         string // O conteúdo completo (para o LLM)
	Summary            string // O sumário curto (ex: "2 arquivos lidos", "Testes passados")
	RequiresFullOutput bool   // Se true, a TUI deve exibir o FullResult mesmo fora do modo verboso (ex: erro de teste)
}

// Tool define a interface para todas as ferramentas do agente.
type Tool interface {
	Name() string
	Execute(args map[string]interface{}) (ToolResult, error)
	SummarizeArgs(args map[string]interface{}) string
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

type Edit struct {
	OldString  string `json:"oldString"`
	NewString  string `json:"newString"`
	ReplaceAll bool   `json:"replaceAll"`
}

type EditArgs struct {
	Path  string `json:"path"`
	Edits []Edit `json:"edits"`
}

// GitToolArgs defines the arguments for the git_tool.
type GitToolArgs struct {
	Action  string   `json:"action"`  // diff, add, commit
	Files   []string `json:"files"`   // for add
	Message string   `json:"message"` // for commit
}

// GoToolArgs defines the arguments for the go_tool.
type GoToolArgs struct {
	Action  string `json:"action"`  // doc, test
	Package string `json:"package"` // for doc
	Path    string `json:"path"`    // for test
}

// ListFilesArgs defines the arguments for the list_files tool.
type ListFilesArgs struct {
	Path      string `json:"path"`
	Pattern   string `json:"pattern"`   // Glob pattern
	Recursive bool   `json:"recursive"` // If true, search recursively
}

// MoveFileArgs defines the arguments for the move_file tool.
type MoveFileArgs struct {
	OldPath string `json:"old_path"`
	NewPath string `json:"new_path"`
}

// Existing Args for backward compatibility during refactor
type GitAddArgs struct {
	Files []string `json:"files"`
}

type GitCommitArgs struct {
	Message string `json:"message"`
}

type GoDocArgs struct {
	Package string `json:"package"`
}

type GoTestArgs struct {
	Path string `json:"path"`
}

// mapToStruct is a helper to decode map into a struct using JSON tags.
func mapToStruct(m map[string]interface{}, s interface{}) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, s)
}
