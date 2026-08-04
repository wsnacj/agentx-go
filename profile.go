package agentx

// ExecutionProfile 描述一次执行的六个正交维度。
//
// 当前根 Client 只接受 DefaultExecutionProfile 返回的同步 open-tool-loop 画像。
// 这些字符串是可观测合同，不表示调用方可任意组合底层 Runtime 能力。
type ExecutionProfile struct {
	Activation         string
	ControlMode        string
	ExecutionIntensity string
	Driver             string
	ResultPolicy       string
	Lifecycle          string
}

// DefaultExecutionProfile 返回根 Client 当前唯一支持的执行画像。
//
// Config.Profile 的零值等价于该画像。需要 Workflow、Objective、Resume 或长任务
// 组合的调用方应使用对应 Runtime Host Kit，不应修改这些字段来伪装启用能力。
func DefaultExecutionProfile() ExecutionProfile {
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
		return DefaultExecutionProfile(), nil
	}
	if profile != DefaultExecutionProfile() {
		return ExecutionProfile{}, newError(CodeUnsupportedProfile, nil)
	}
	return profile, nil
}
