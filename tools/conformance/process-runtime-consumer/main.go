package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	process "github.com/wsnacj/agentx-go/tools/process"
)

type result struct {
	Output        string `json:"output"`
	ExitCode      int    `json:"exit_code"`
	Truncated     bool   `json:"truncated"`
	TimeoutStatus string `json:"timeout_status"`
	Verified      bool   `json:"verified"`
}

func run(ctx context.Context) (result, error) {
	root, err := os.MkdirTemp("", "agentx-process-consumer-")
	if err != nil {
		return result{}, err
	}
	defer os.RemoveAll(root)
	adapter := process.NewLocalAdapter(process.LocalOptions{Root: root, MaxOutputBytes: 16})
	success, err := adapter.Run(ctx, process.Command{Command: "printf agentx-process-ready"})
	if err != nil {
		return result{}, err
	}
	timed, err := adapter.Run(ctx, process.Command{Command: "while :; do :; done", Timeout: 30 * time.Millisecond})
	if err != nil {
		return result{}, err
	}
	value := result{
		Output: success.Stdout, ExitCode: success.ExitCode,
		Truncated: success.StdoutTruncated, TimeoutStatus: timed.Status,
	}
	value.Verified = value.Output == "agentx-process-r" && value.ExitCode == 0 &&
		value.Truncated && value.TimeoutStatus == "timed_out"
	return value, nil
}

func main() {
	value, err := run(context.Background())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if !value.Verified || !strings.HasPrefix(value.Output, "agentx-") {
		fmt.Fprintln(os.Stderr, "process runtime conformance failed")
		os.Exit(1)
	}
	if err := json.NewEncoder(os.Stdout).Encode(value); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
