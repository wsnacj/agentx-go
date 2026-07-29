package main

import (
	"encoding/json"
	"errors"
	"fmt"

	agentxprotocol "github.com/wsnacj/agentx-go/runtime/protocol"
	agentxsafeerror "github.com/wsnacj/agentx-go/runtime/telemetry/safeerror"
)

func canonicalEventJSON() ([]byte, error) {
	event := agentxprotocol.NormalizeRunEvent(agentxprotocol.RunEvent{
		Envelope: agentxprotocol.Envelope{
			Kind:  " tool.call.completed ",
			RunID: " run-consumer-1 ",
		},
		Status:   " Completed ",
		ToolName: " Browser_Open ",
	})
	if err := agentxprotocol.ValidateRunEvent(event); err != nil {
		return nil, err
	}
	return json.Marshal(event)
}

func canonicalSafeErrorJSON() ([]byte, error) {
	cause := errors.New("private consumer sentinel")
	wrapped := agentxsafeerror.WrapWithIdentity(
		cause,
		"operation failed",
		"consumer-error-1",
	)
	if !errors.Is(wrapped, cause) {
		return nil, fmt.Errorf("safeerror wrapper lost cause")
	}
	projection := agentxsafeerror.Project(
		wrapped,
		" Runtime Error ",
		" UPSTREAM/FAILED ",
	)
	return json.Marshal(projection)
}

func main() {
	payload, err := canonicalEventJSON()
	if err != nil {
		panic(err)
	}
	fmt.Println(string(payload))
}
