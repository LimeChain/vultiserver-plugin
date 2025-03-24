package types

import "encoding/json"

type PluginPricingCreateDto struct {
	PublicKey  string          `json:"public_key" validate:"required"`
	PluginType string          `json:"plugin_type" validate:"required"`
	Signature  string          `json:"signature" validate:"required"`
	Pricing    json.RawMessage `json:"pricing" validate:"required"`
}

type PluginPricing struct {
	ID         string          `json:"id" validate:"required"`
	PublicKey  string          `json:"public_key" validate:"required"`
	PluginType string          `json:"plugin_type" validate:"required"`
	Signature  string          `json:"signature" validate:"required"`
	Pricing    json.RawMessage `json:"pricing" validate:"required"`
}
