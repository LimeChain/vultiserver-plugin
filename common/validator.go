package common

import (
	"encoding/hex"
	"fmt"

	"github.com/vultisig/vultisigner/internal/types"
)

/**
 * Basic plugin pricing validator to be reused across services.
 * Verifier should also compare signed pricing matches general definition
 * in plugin->policy for the plugin it is signed for.
 */
func ValidatePluginPricingPolicy(pricing *types.PluginPricing, pluginType string) error {
	if pricing.PluginType != pluginType {
		return fmt.Errorf("pricing policy does not match plugin type, expected: %s, got: %s", pluginType, pricing.PluginType)
	}

	if pricing.ChainCodeHex == "" {
		return fmt.Errorf("pricing policy does not contain chain_code_hex")
	}

	if pricing.PublicKey == "" {
		return fmt.Errorf("pricing policy does not contain public_key")
	}

	pubKeyBytes, err := hex.DecodeString(pricing.PublicKey)
	if err != nil {
		return fmt.Errorf("invalid hex encoding: %w", err)
	}

	isValidPublicKey := CheckIfPublicKeyIsValid(pubKeyBytes, pricing.IsEcdsa)

	if !isValidPublicKey {
		return fmt.Errorf("invalid public_key")
	}

	return nil
}
