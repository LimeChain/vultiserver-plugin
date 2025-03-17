package types

import "encoding/json"

type Plugin struct {
	ID             string          `json:"id" validate:"required"`
	Title          string          `json:"title" validate:"required"`
	Description    string          `json:"description" validate:"required"`
	Metadata       json.RawMessage `json:"metadata" validate:"required"`
	ServerEndpoint string          `json:"server_endpoint" validate:"required"`
	PricingID      string          `json:"pricing_id" validate:"required"`
}

type PluginCreateDto struct {
	Title          string          `json:"title" validate:"required"`
	Description    string          `json:"description" validate:"required"`
	Metadata       json.RawMessage `json:"metadata" validate:"required"`
	ServerEndpoint string          `json:"server_endpoint" validate:"required"`
	PricingID      string          `json:"pricing_id" validate:"required"`
}
