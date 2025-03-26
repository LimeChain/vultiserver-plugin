package plugin

import (
	"fmt"
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/vultisig/vultisigner/pkg/uniswap"
)

type RawTxData struct {
	TxHash     []byte
	RlpTxBytes []byte
	Type       string
}

func GenerateFeeTransaction(uniswapClient *uniswap.Client, chainID *big.Int, signerAddress *common.Address, amount float64) (*RawTxData, error) {
	// TODO: check allowance?

	// TODO: from config?
	pluginFeeWallet := common.HexToAddress("0xa1172EcfaB3f313d0f0975717106a04626E4d485")

	nonceOffset := uint64(0)
	txHash, rawTx, err := uniswapClient.ERC20Transfer(
		chainID,
		signerAddress,
		&pluginFeeWallet,
		big.NewInt(int64(math.Round(amount*10000))),
		nonceOffset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to make FEE transaction: %w", err)
	}

	return &RawTxData{txHash, rawTx, "FEE"}, nil
}
