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

func GenerateFeeTransactions(
	uniswapClient *uniswap.Client,
	chainID *big.Int,
	signerAddress *common.Address,
	feeTokenAddress *common.Address,
	amount float64,
) ([]RawTxData, error) {
	// TODO: from config?
	pluginFeeWallet := common.HexToAddress("0xa1172EcfaB3f313d0f0975717106a04626E4d485")

	nonceOffset := uint64(0)
	var rawTxsData []RawTxData

	parsedAmount := big.NewInt(int64(math.Round(amount * 10000))) // TODO: if percentage/absolute, if token precision 6 etc..

	// txHash, rawTx, err := uniswapClient.ApproveERC20Token(
	// 	chainID,
	// 	signerAddress,
	// 	*feeTokenAddress,
	// 	pluginFeeWallet,
	// 	parsedAmount,
	// 	nonceOffset,
	// )
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to make approve FEE transaction: %w", err)
	// }

	// approveFeeTx := &RawTxData{txHash, rawTx, "APPROVE_FEE"}
	// rawTxsData = append(rawTxsData, approveFeeTx)

	// nonceOffset++

	txHash, rawTx, err := uniswapClient.ERC20Transfer(
		chainID,
		feeTokenAddress,
		signerAddress,
		&pluginFeeWallet,
		parsedAmount,
		nonceOffset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to make FEE transaction: %w", err)
	}

	feeTx := RawTxData{txHash, rawTx, "FEE"}
	rawTxsData = append(rawTxsData, feeTx)

	return rawTxsData, nil
}
