package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

const (
	shellTimeout   = 30 * time.Second
	shellOutputCap = 10 * 1024 // 10KB
)

type shellExecArgs struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

// ShellExec returns a ToolFunc that executes shell commands within the project
// directory. Any command may be run; only timeout and output cap are enforced.
func ShellExec(baseDir string) ToolFunc {
	return func(ctx context.Context, raw json.RawMessage) (string, error) {
		var args shellExecArgs
		if err := json.Unmarshal(raw, &args); err != nil {
			return "", fmt.Errorf("invalid args: %w", err)
		}
		if args.Command == "" {
			return "", fmt.Errorf("command is required")
		}

		ctx, cancel := context.WithTimeout(ctx, shellTimeout)
		defer cancel()

		cmd := exec.CommandContext(ctx, args.Command, args.Args...)
		cmd.Dir = baseDir
		if dir := ProjectDirFromContext(ctx); dir != "" {
			cmd.Dir = dir
		}

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()

		out := stdout.String()
		if len(out) > shellOutputCap {
			out = out[:shellOutputCap] + "\n... output truncated"
		}
		errOut := stderr.String()
		if len(errOut) > shellOutputCap {
			errOut = errOut[:shellOutputCap] + "\n... output truncated"
		}

		result := out
		if errOut != "" {
			result += "\nSTDERR:\n" + errOut
		}

		if err != nil {
			return result, fmt.Errorf("command failed: %w", err)
		}
		return result, nil
	}
}
