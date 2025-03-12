package types

import "encoding/json"

type Plugin struct {
	ID             string          `json:"id" validate:"required"`
	Title          string          `json:"title" validate:"required"`
	Description    string          `json:"description"`
	Metadata       json.RawMessage `json:"metadata"`
	ServerEndpoint string          `json:"server_endpoint"`
	PricingID      string          `json:"pricing_id"`
}

type PluginCreateDto struct {
	Title          string          `json:"title" validate:"required"`
	Description    string          `json:"description"`
	Metadata       json.RawMessage `json:"metadata"`
	ServerEndpoint string          `json:"server_endpoint"`
	PricingID      string          `json:"pricing_id"` // TODO: pricing struct instead so pricing and plugin are created with 1 request?
}
