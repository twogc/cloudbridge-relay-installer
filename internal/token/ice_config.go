package token

import (
	"fmt"
	"time"
)

// IceConfig represents ICE configuration
type IceConfig struct {
	StunServers       []string `json:"stun_servers"`
	TurnServers       []string `json:"turn_servers"`
	IceTimeout        string   `json:"ice_timeout"`
	ConnectionTimeout string   `json:"connection_timeout"`
	GatheringTimeout  string   `json:"gathering_timeout"`
	CandidateTypes    []string `json:"candidate_types"`
}

// Validate validates ICE configuration
func (i *IceConfig) Validate() error {
	if len(i.StunServers) == 0 {
		return fmt.Errorf("at least one STUN server is required")
	}

	if i.IceTimeout != "" {
		if _, err := time.ParseDuration(i.IceTimeout); err != nil {
			return fmt.Errorf("invalid ice_timeout: %v", err)
		}
	}

	if i.ConnectionTimeout != "" {
		if _, err := time.ParseDuration(i.ConnectionTimeout); err != nil {
			return fmt.Errorf("invalid connection_timeout: %v", err)
		}
	}

	if i.GatheringTimeout != "" {
		if _, err := time.ParseDuration(i.GatheringTimeout); err != nil {
			return fmt.Errorf("invalid gathering_timeout: %v", err)
		}
	}

	validCandidateTypes := map[string]bool{
		"host":  true,
		"srflx": true,
		"relay": true,
		"prflx": true,
	}

	for _, candidateType := range i.CandidateTypes {
		if !validCandidateTypes[candidateType] {
			return fmt.Errorf("invalid candidate_type: %s", candidateType)
		}
	}

	return nil
}

