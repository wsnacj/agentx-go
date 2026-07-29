package main

import (
	"encoding/json"
	"fmt"

	agentxprotocol "github.com/wsnacj/agentx-go/runtime/protocol"
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

func main() {
	payload, err := canonicalEventJSON()
	if err != nil {
		panic(err)
	}
	fmt.Println(string(payload))
}
