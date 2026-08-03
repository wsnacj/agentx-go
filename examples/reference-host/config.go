package main

import (
	"flag"
	"fmt"
	"io"
	"strings"
)

const (
	modeChat     = "chat"
	modeToolLoop = "tool-loop"

	providerFixture = "fixture"
	toolsNone       = "none"
	toolsDiffs      = "diffs"
	storeMemory     = "memory"
)

type config struct {
	Mode     string
	Provider string
	Tools    string
	Store    string
	Input    string
}

func parseConfig(arguments []string, output io.Writer) (config, error) {
	value := config{
		Mode:     modeChat,
		Provider: providerFixture,
		Tools:    toolsNone,
		Store:    storeMemory,
		Input:    "hello AgentX",
	}
	flags := flag.NewFlagSet("agentx-reference-host", flag.ContinueOnError)
	flags.SetOutput(output)
	flags.StringVar(&value.Mode, "mode", value.Mode, "chat or tool-loop")
	flags.StringVar(&value.Provider, "provider", value.Provider, "fixture")
	flags.StringVar(&value.Tools, "tools", value.Tools, "none or diffs")
	flags.StringVar(&value.Store, "store", value.Store, "memory")
	flags.StringVar(&value.Input, "input", value.Input, "run input")
	if err := flags.Parse(arguments); err != nil {
		return config{}, err
	}
	return normalizeConfig(value)
}

func normalizeConfig(value config) (config, error) {
	value.Mode = strings.TrimSpace(value.Mode)
	value.Provider = strings.TrimSpace(value.Provider)
	value.Tools = strings.TrimSpace(value.Tools)
	value.Store = strings.TrimSpace(value.Store)
	value.Input = strings.TrimSpace(value.Input)
	if value.Mode != modeChat && value.Mode != modeToolLoop {
		return config{}, fmt.Errorf("reference host: unsupported mode %q", value.Mode)
	}
	if value.Provider != providerFixture {
		return config{}, fmt.Errorf("reference host: unsupported provider %q", value.Provider)
	}
	if value.Tools != toolsNone && value.Tools != toolsDiffs {
		return config{}, fmt.Errorf("reference host: unsupported tools %q", value.Tools)
	}
	if value.Store != storeMemory {
		return config{}, fmt.Errorf("reference host: unsupported store %q", value.Store)
	}
	if value.Input == "" {
		return config{}, fmt.Errorf("reference host: input is required")
	}
	if value.Mode == modeChat && value.Tools != toolsNone {
		return config{}, fmt.Errorf("reference host: chat mode requires tools=none")
	}
	if value.Mode == modeToolLoop && value.Tools != toolsDiffs {
		return config{}, fmt.Errorf("reference host: tool-loop mode requires tools=diffs")
	}
	return value, nil
}
