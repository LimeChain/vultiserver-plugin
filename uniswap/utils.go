package uniswap

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func ToKeyPairAndAddress(privateKeyHex string) (*ecdsa.PrivateKey, *ecdsa.PublicKey, common.Address, error) {
	if privateKeyHex == "" {
		return nil, nil, common.Address{}, fmt.Errorf("private key is not set")
	}
	ecdsaPrivateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, nil, common.Address{}, fmt.Errorf("failed to get private key: %v", err)
	}
	ecdsaPublicKey := ecdsaPrivateKey.Public().(*ecdsa.PublicKey)

	address := crypto.PubkeyToAddress(*ecdsaPublicKey)

	return ecdsaPrivateKey, ecdsaPublicKey, address, nil
}

func ToPublicKeyAndAddress(pubKeyHex string) (*ecdsa.PublicKey, *common.Address, error) {
	if pubKeyHex == "" {
		return nil, nil, fmt.Errorf("public key is not set")
	}
	data, err := hex.DecodeString(pubKeyHex)
	if err != nil {
		return nil, nil, err
	}

	var ecdsaPubKey *ecdsa.PublicKey

	if len(data) == 33 {
		ecdsaPubKey, err = crypto.DecompressPubkey(data)
		if err != nil {
			return nil, nil, err
		}
	}

	if len(data) == 65 {
		ecdsaPubKey, err = crypto.UnmarshalPubkey(data)
		if err != nil {
			return nil, nil, err
		}
	}

	address := crypto.PubkeyToAddress(*ecdsaPubKey)

	return ecdsaPubKey, &address, nil
}
