package types

import "encoding/json"

type PluginPricingCreateDto struct {
	PublicKey    string          `json:"public_key" validate:"required"`
	PluginType   string          `json:"plugin_type" validate:"required"`
	IsEcdsa      bool            `json:"is_ecdsa" validate:"required"`
	ChainCodeHex string          `json:"chain_code_hex" validate:"required"`
	DerivePath   string          `json:"derive_path" validate:"required"`
	Signature    string          `json:"signature" validate:"required"`
	Pricing      json.RawMessage `json:"pricing" validate:"required"`
}

type PluginPricing struct {
	ID           string          `json:"id" validate:"required"`
	PublicKey    string          `json:"public_key" validate:"required"`
	PluginType   string          `json:"plugin_type" validate:"required"`
	IsEcdsa      bool            `json:"is_ecdsa" validate:"required"`
	ChainCodeHex string          `json:"chain_code_hex" validate:"required"`
	DerivePath   string          `json:"derive_path" validate:"required"`
	Signature    string          `json:"signature" validate:"required"`
	Pricing      json.RawMessage `json:"pricing" validate:"required"`
}

type PricingPolicy struct {
	Type string `json:"type"`
	// Frequency string  `json:"frequency"`
	Amount float64 `json:"amount"`
	Metric string  `json:"metric"`
}
