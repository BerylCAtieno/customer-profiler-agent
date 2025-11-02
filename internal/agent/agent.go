package agent

import (
	_ "embed"
	"fmt"
)

//go:embed agent.json
var AgentCardData []byte

func LoadAgentCard() error {
	if len(AgentCardData) == 0 {
		return fmt.Errorf("agent card is empty")
	}
	return nil
}

// TODO: update times in json
// TODO: update domains
