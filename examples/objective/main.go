package main

import (
	"context"
	"fmt"

	objective "github.com/wsnacj/agentx-go/runtime/objective"
	objectivehostkit "github.com/wsnacj/agentx-go/runtime/objective/hostkit"
)

func run(ctx context.Context) (string, error) {
	runtime, err := objectivehostkit.New(objectivehostkit.Config{
		Handlers: map[objective.DisplaySafeRef]objectivehostkit.Handler{
			"adapter:conformance": execute,
		},
	})
	if err != nil {
		return "", err
	}
	result := runtime.Run(ctx, objectivehostkit.RunRequest{
		Ingress:               ingress(),
		DispatchEnabled:       true,
		DispatchHostConfirmed: true,
	})
	if !result.Completed {
		return "", fmt.Errorf("objective blocked: status=%s failure=%s next=%s", result.Status, result.FailureClass, result.NextHostAction)
	}
	return fmt.Sprintf("agentx-objective-hostkit-ok:%s:%s:%t", result.Status, result.Dispatch.AdapterRef, result.Dispatch.HostExecutionReported), nil
}

func main() {
	output, err := run(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Println(output)
}
