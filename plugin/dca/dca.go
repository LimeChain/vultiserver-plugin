package dca

import (
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"strings"

	gcommon "github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/vultisig/vultisigner/common"
	"github.com/vultisig/vultisigner/internal/types"
	"github.com/vultisig/vultisigner/storage"
	"github.com/vultisig/vultisigner/uniswap"
)

const (
	PLUGIN_TYPE    = "dca"
	PLUGIN_VERSION = "0.0.1"
	POLICY_VERSION = "0.0.1"
)

type DCAPlugin struct {
	uniswapClient *uniswap.Client
	db            storage.DatabaseStorage
	logger        *logrus.Logger
}

func NewDCAPlugin(uniswapCfg *uniswap.Config, db storage.DatabaseStorage, logger *logrus.Logger) *DCAPlugin {
	uniswapClient, err := uniswap.NewClient(uniswapCfg)
	if err != nil {
		log.Fatalf("Failed to initialize Uniswap client: %v", err)
	}

	return &DCAPlugin{uniswapClient, db, logger}
}

func (p *DCAPlugin) SignPluginMessages(e echo.Context) error {
	p.logger.Warn("DCA: SIGN PLUGIN MESSAGES")
	return nil
}

func (p *DCAPlugin) SetupPluginPolicy(policyDoc *types.PluginPolicy) error {
	if policyDoc == nil {
		return fmt.Errorf("no policy to set up")
	}

	if policyDoc.ID == "" {
		policyDoc.ID = uuid.NewString()
	}

	if policyDoc.PolicyVersion == "" {
		policyDoc.PolicyVersion = POLICY_VERSION
	}

	if policyDoc.PluginVersion == "" {
		policyDoc.PluginVersion = PLUGIN_VERSION
	}

	if policyDoc.PluginID == "" {
		policyDoc.PluginID = uuid.NewString()
	}

	return nil
}

func (p *DCAPlugin) ValidatePluginPolicy(policyDoc types.PluginPolicy) error {
	if policyDoc.PluginType != PLUGIN_TYPE {
		return fmt.Errorf("policy does not match plugin type, expected: %s, got: %s", PLUGIN_TYPE, policyDoc.PluginType)
	}

	var dcaPolicy types.DCAPolicy
	if err := json.Unmarshal(policyDoc.Policy, &dcaPolicy); err != nil {
		return fmt.Errorf("failed to unmarshal DCA policy: %w", err)
	}

	mixedCaseTokenIn, err := gcommon.NewMixedcaseAddressFromString(dcaPolicy.SourceTokenID)
	if err != nil {
		return fmt.Errorf("invalid source token address: %s", dcaPolicy.SourceTokenID)
	}
	if strings.ToLower(dcaPolicy.SourceTokenID) != dcaPolicy.SourceTokenID {
		if !mixedCaseTokenIn.ValidChecksum() {
			return fmt.Errorf("invalid source token address checksum: %s", dcaPolicy.SourceTokenID)
		}
	}

	mixedCaseTokenOut, err := gcommon.NewMixedcaseAddressFromString(dcaPolicy.DestinationTokenID)
	if err != nil {
		return fmt.Errorf("invalid destination token address: %s", dcaPolicy.DestinationTokenID)
	}
	if strings.ToLower(dcaPolicy.DestinationTokenID) != dcaPolicy.DestinationTokenID {
		if !mixedCaseTokenOut.ValidChecksum() {
			return fmt.Errorf("invalid destination token address checksum: %s", dcaPolicy.DestinationTokenID)
		}
	}

	if dcaPolicy.SourceTokenID == dcaPolicy.DestinationTokenID {
		return fmt.Errorf("source token and destination token addresses are the same")
	}

	if dcaPolicy.TotalAmount == "" {
		return fmt.Errorf("total amount is required")
	}
	totalAmount, ok := new(big.Int).SetString(dcaPolicy.TotalAmount, 10)
	if !ok {
		return fmt.Errorf("invalid total amount %s", dcaPolicy.TotalAmount)
	}
	if totalAmount.Cmp(big.NewInt(0)) <= 0 {
		return fmt.Errorf("total amount must be greater than 0")
	}

	if dcaPolicy.TotalOrders == "" {
		return fmt.Errorf("total orders is required")
	}
	totalOrders, ok := new(big.Int).SetString(dcaPolicy.TotalOrders, 10)
	if !ok {
		return fmt.Errorf("invalid total orders %s", dcaPolicy.TotalOrders)
	}
	if totalOrders.Cmp(big.NewInt(0)) <= 0 {
		return fmt.Errorf("total orders must be greater than 0")
	}

	if dcaPolicy.PriceRange.Min != "" && dcaPolicy.PriceRange.Max != "" && dcaPolicy.PriceRange.Min >= dcaPolicy.PriceRange.Max {
		return fmt.Errorf("min price range should be equal or lower than max price range")
	}

	// if dcaPolicy.SlippagePercentage == "" {
	// 	return fmt.Errorf("slippage percentage is required")
	// }
	// slippage, err := strconv.ParseFloat(dcaPolicy.SlippagePercentage, 64)
	// if err != nil {
	// 	return fmt.Errorf("invalid slippage percentage %s", dcaPolicy.SlippagePercentage)
	// }
	// if slippage <= 0 || slippage > 100 {
	// 	return fmt.Errorf("slippage percentage must be between 0 and 100 %s", dcaPolicy.SlippagePercentage)
	// }

	if dcaPolicy.ChainID == "" {
		return fmt.Errorf("chain id is required")
	}

	return nil
}

func (p *DCAPlugin) ConfigurePlugin(e echo.Context) error {
	return nil
}

func (p *DCAPlugin) ProposeTransactions(policy types.PluginPolicy) ([]types.PluginKeysignRequest, error) {
	p.logger.Warn("DCA: PROPOSE TRANSACTIONS")

	var txs []types.PluginKeysignRequest

	// validate policy
	err := p.ValidatePluginPolicy(policy)
	if err != nil {
		return txs, fmt.Errorf("failed to validate plugin policy: %v", err)
	}

	var dcaPolicy types.DCAPolicy
	if err := json.Unmarshal(policy.Policy, &dcaPolicy); err != nil {
		return txs, fmt.Errorf("fail to unmarshal dca policy, err: %w", err)
	}
	// build transactions
	// TODO: get theses values from the vault
	derivePath := "m/44'/60'/0'/0/0"                                                       // ethereum
	hexChainCode := "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"     // vault's chain code
	hexEncryptionKey := "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef" // vault's encryption key
	vaultPassword := "your-secure-password"                                                // vault's password

	signerAddress, err := common.DeriveAddress(policy.PublicKey, hexChainCode, derivePath)
	if err != nil {
		return []types.PluginKeysignRequest{}, fmt.Errorf("failed to derive address: %v", err)
	}

	chainID, ok := big.NewInt(0).SetString(dcaPolicy.ChainID, 10)
	if !ok {
		return []types.PluginKeysignRequest{}, fmt.Errorf("failed to parse chain id: %v", err)
	}

	rawTxsData, err := p.generateSwapTransactions(chainID, signerAddress, dcaPolicy.SourceTokenID, dcaPolicy.DestinationTokenID, dcaPolicy.TotalAmount)
	if err != nil {
		return []types.PluginKeysignRequest{}, fmt.Errorf("failed to generate transaction hash: %v", err)
	}

	for _, data := range rawTxsData {
		// Create signing request
		signRequest := types.PluginKeysignRequest{
			KeysignRequest: types.KeysignRequest{
				PublicKey:        policy.PublicKey,
				Messages:         []string{hex.EncodeToString(data.TxHash)},
				SessionID:        uuid.New().String(),
				HexEncryptionKey: hexEncryptionKey,
				DerivePath:       derivePath,
				IsECDSA:          true, // TODO
				VaultPassword:    vaultPassword,
				StartSession:     false,
				Parties:          []string{common.PluginPartyID, common.VerifierPartyID},
			},
			Transaction: hex.EncodeToString(data.RlpTxBytes),
			PluginID:    policy.PluginID,
			PolicyID:    policy.ID,
		}
		txs = append(txs, signRequest)
	}

	return txs, nil
}

type RawTxData struct {
	TxHash     []byte
	RlpTxBytes []byte
}

func (p *DCAPlugin) generateSwapTransactions(chainID *big.Int, signerAddress *gcommon.Address, srcToken, destToken string, strAmount string) ([]RawTxData, error) {
	srcTokenAddress := gcommon.HexToAddress(srcToken)
	destTokenAddress := gcommon.HexToAddress(destToken)
	tokensPair := []gcommon.Address{srcTokenAddress, destTokenAddress}

	amount, ok := new(big.Int).SetString(strAmount, 10)
	if !ok {
		return []RawTxData{}, fmt.Errorf("failed to parse amount")
	}
	log.Print("Amount: ", amount.String())

	// fetch token pair amount out
	expectedAmount, err := p.uniswapClient.GetExpectedAmountOut(amount, tokensPair)
	if err != nil {
		log.Fatalf("Failed to get expected amount out: %v", err)
	}
	log.Println("Expected amount out:", expectedAmount.String())

	// TODO: validate the price range (if specified)

	// calculate output amount with slippage
	amountOutMin := p.uniswapClient.CalculateAmountOutMin(expectedAmount, 1.0)

	// TODO: mint + approve once, during policy configuration,
	// so there will be only one transaction to sign each time
	rawTxsData := []RawTxData{}

	// mint WETH
	log.Println("Minting WETH...")
	logTokenBalances(p.uniswapClient, signerAddress, srcTokenAddress, destTokenAddress)
	txHash, rawTx, err := p.uniswapClient.MintWETH(chainID, signerAddress, amount, srcTokenAddress)
	if err != nil {
		log.Fatalf("Failed to mint WETH: %v", err)
	}
	logTokenBalances(p.uniswapClient, signerAddress, srcTokenAddress, destTokenAddress)
	rawTxsData = append(rawTxsData, RawTxData{txHash, rawTx})

	// approve Router to spend input token
	log.Printf("Approving Uniswap Router to spend %s...", srcTokenAddress.Hex())
	txHash, rawTx, err = p.uniswapClient.ApproveERC20Token(chainID, signerAddress, srcTokenAddress, *p.uniswapClient.GetRouterAddress(), amount)
	if err != nil {
		log.Fatalf("Failed to approve token: %v", err)
	}
	rawTxsData = append(rawTxsData, RawTxData{txHash, rawTx})

	// swap tokens
	txHash, rawTx, err = p.uniswapClient.SwapTokens(chainID, signerAddress, amount, amountOutMin, tokensPair)
	if err != nil {
		log.Fatalf("Failed to swap tokens: %v", err)
	}
	logTokenBalances(p.uniswapClient, signerAddress, srcTokenAddress, destTokenAddress)
	rawTxsData = append(rawTxsData, RawTxData{txHash, rawTx})

	return rawTxsData, nil
}

func (p *DCAPlugin) ValidateTransactionProposal(policy types.PluginPolicy, txs []types.PluginKeysignRequest) error {
	p.logger.Warn("DCA: VALIDATE TRANSACTION PROPOSAL")
	return nil
}

func (p *DCAPlugin) Frontend() embed.FS {
	return embed.FS{}
}

func logTokenBalances(client *uniswap.Client, signerAddress *gcommon.Address, tokenInAddress, tokenOutAddress gcommon.Address) {
	tokenInBalance, err := client.GetTokenBalance(signerAddress, tokenInAddress)
	if err != nil {
		log.Printf("Error getting input token balance: %v", err)
		return
	}
	log.Printf("input token balance: %s", tokenInBalance.String())

	tokenOutBalance, err := client.GetTokenBalance(signerAddress, tokenOutAddress)
	if err != nil {
		log.Printf("Error getting output token balance: %v", err)
		return
	}
	log.Printf("output token balance: %s", tokenOutBalance.String())
}
