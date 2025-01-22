package main

import (
	"context"
	"flag"
	"fmt"
	"math/big"
	"os"
	"path/filepath"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/vultisig/vultisigner/config"
	"github.com/vultisig/vultisigner/uniswap"
)

var (
	privateKey = os.Getenv("PRIVATE_KEY")
)

var vaultName string
var stateDir string

func main() {
	flag.StringVar(&vaultName, "vault", "", "vault name")
	flag.StringVar(&stateDir, "state-dir", "", "state directory")
	flag.Parse()

	if vaultName == "" {
		panic("vault name is required")
	}

	if stateDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			panic(err)
		}

		stateDir = filepath.Join(homeDir, ".vultiserver", "vaults")
	}

	keyPath := filepath.Join("vaults", vaultName, "public_key")
	rawKeyBytes, err := os.ReadFile(keyPath)
	if err != nil {
		panic(err)
	}
	hexKey := string(rawKeyBytes)

	_, vaultAddress, err := uniswap.ToPublicKeyAndAddress(hexKey)
	if err != nil {
		panic(err)
	}
	fmt.Println("To vault address:", vaultAddress.Hex())

	pluginConfig, err := config.ReadConfig("config-plugin")
	if err != nil {
		panic(err)
	}

	rpcClient, err := ethclient.Dial(pluginConfig.Server.Plugin.Eth.Rpc)
	if err != nil {
		panic(err)
	}

	signerPrivateKey, _, signerAddress, err := uniswap.ToKeyPairAndAddress(privateKey)
	if err != nil {
		panic(err)
	}
	// 1 eth
	amount := big.NewInt(1000000000000000000)

	gasPrice, err := rpcClient.SuggestGasPrice(context.Background())
	if err != nil {
		panic(err)
	}
	nonce, err := rpcClient.PendingNonceAt(context.Background(), signerAddress)
	if err != nil {
		panic(err)
	}
	gasLimit, err := rpcClient.EstimateGas(context.Background(), ethereum.CallMsg{})
	if err != nil {
		panic(err)
	}
	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		To:       vaultAddress,
		Value:    amount,
		Gas:      gasLimit,
		GasPrice: gasPrice,
		Data:     nil,
	})

	chainID, err := rpcClient.NetworkID(context.Background())
	if err != nil {
		panic(err)
	}
	signer := types.NewEIP155Signer(chainID)
	signedTx, err := types.SignTx(tx, signer, signerPrivateKey)
	if err != nil {
		panic(err)
	}

	sender, err := signer.Sender(signedTx)
	if err != nil {
		panic("failed to get sender: " + err.Error())
	}
	fmt.Println("From sender address:", sender.Hex())

	err = rpcClient.SendTransaction(context.Background(), signedTx)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Transaction sent: %s", signedTx.Hash().Hex())

	receipt, err := bind.WaitMined(context.Background(), rpcClient, signedTx)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Transaction receipt status: %v", receipt.Status)
}
