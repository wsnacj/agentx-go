package agentx

// ExecutionProfile 描述一次执行的六个正交维度。
//
// W1 只接受同步 open-tool-loop 画像。这些字符串是已测量合同，不表示调用方可通过
// 任意组合重新配置底层 Runtime。
type ExecutionProfile struct {
	Activation         string
	ControlMode        string
	ExecutionIntensity string
	Driver             string
	ResultPolicy       string
	Lifecycle          string
}

func defaultProfile() ExecutionProfile {
	return ExecutionProfile{
		Activation:         "off",
		ControlMode:        "tool",
		ExecutionIntensity: "l2_bounded_tool_loop",
		Driver:             "open_tool_loop",
		ResultPolicy:       "runner_final_reply",
		Lifecycle:          "synchronous_run",
	}
}

func resolveProfile(profile ExecutionProfile) (ExecutionProfile, error) {
	if profile == (ExecutionProfile{}) {
		return defaultProfile(), nil
	}
	if profile != defaultProfile() {
		return ExecutionProfile{}, newError(CodeUnsupportedProfile, nil)
	}
	return profile, nil
}
