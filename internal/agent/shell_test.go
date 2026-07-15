package agent

import (
	"testing"
)

func TestShellExecutor_Execute(t *testing.T) {
	executor := &ShellExecutor{}

	tests := []struct {
		name    string
		command string
		want    string
		wantErr bool
	}{
		{
			name:    "echo command",
			command: "echo hello",
			want:    "hello",
			wantErr: false,
		},
		{
			name:    "ls command (assuming it exists)",
			command: "ls",
			want:    "", // We don't check content exactly as it's environment dependent, but let's see if it runs
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := executor.Execute(tt.command)
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.name == "echo command" && got != tt.want {
				t.Errorf("Execute() got = %v, want %v", got, tt.want)
			}
		})
	}
}
