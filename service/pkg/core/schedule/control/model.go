package control

import "github.com/scienceol/studio/service/pkg/core/schedule/engine"

type ActionStatus struct {
	engine.ActionKey
	engine.ActionValue
}

type ActionPong struct {
	PingID          string  `json:"ping_id"`
	ClientTimestamp float64 `json:"client_timestamp"`
	ServerTimestamp float64 `json:"server_timestamp"`
}
