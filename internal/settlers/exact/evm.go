package exact

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"
	"x402/internal/cryptohelpers"
	"x402/pkg/x402"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

type evmSettler struct {
	cli      *ethclient.Client // full RPC client
	usdc     common.Address    // token contract
	abi      abi.ABI           // cached ABI
	gasPayer *ecdsa.PrivateKey // facilitator’s key → pays gas
	chainID  *big.Int
}

func NewEIP3009SettlerCore(rpcURL, usdcAddr string, gasKey *ecdsa.PrivateKey) (x402.Settler, error) {
	cli, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, err
	}

	parsed, err := abi.JSON(strings.NewReader(usdcEIP3009ABI))
	if err != nil {
		return nil, err
	}

	id, err := cli.ChainID(context.Background())
	if err != nil {
		return nil, err
	}

	return &evmSettler{
		cli:      cli,
		usdc:     common.HexToAddress(usdcAddr),
		abi:      parsed,
		gasPayer: gasKey,
		chainID:  id,
	}, nil
}

func (s *evmSettler) Settle(ctx context.Context, p *x402.PaymentPayload, req *x402.PaymentRequirements) (txHash string, err error) {
	evm := p.Payload
	auth := evm.Authorization

	v, r32, s32, err := cryptohelpers.SplitSig(evm.Signature)
	if err != nil {
		return "", err
	}

	value, ok := new(big.Int).SetString(auth.Value, 10)
	if !ok {
		return "", fmt.Errorf("invalid auth value")
	}

	validAfter, ok := new(big.Int).SetString(auth.ValidAfter, 10)
	if !ok {
		return "", fmt.Errorf("invalid validAfter")
	}

	validBefore, ok := new(big.Int).SetString(auth.ValidBefore, 10)
	if !ok {
		return "", fmt.Errorf("invalid validBefore")
	}

	nonce32, err := cryptohelpers.HexTo32(auth.Nonce)
	if err != nil {
		return "", fmt.Errorf("could not parse nonce " + err.Error())
	}

	data, err := s.abi.Pack(
		"transferWithAuthorization",
		common.HexToAddress(auth.From),
		common.HexToAddress(auth.To),
		value,
		validAfter,
		validBefore,
		nonce32,
		v, r32, s32,
	)
	if err != nil {
		return "", err
	}

	fromAddr := crypto.PubkeyToAddress(s.gasPayer.PublicKey)
	nonce, err := s.cli.PendingNonceAt(ctx, fromAddr)
	if err != nil {
		return "", err
	}

	gasTipCap, err := s.cli.SuggestGasTipCap(ctx)
	if err != nil {
		return "", err
	}

	gasFeeCap, err := s.cli.SuggestGasPrice(ctx)
	if err != nil {
		return "", err
	}

	msg := ethereum.CallMsg{From: fromAddr, To: &s.usdc, Data: data}

	gasLimit, err := s.cli.EstimateGas(ctx, msg)
	if err != nil {
		return "", err
	}

	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   s.chainID,
		Nonce:     nonce,
		GasTipCap: gasTipCap,
		GasFeeCap: gasFeeCap,
		Gas:       gasLimit,
		To:        &s.usdc,
		Data:      data,
	})

	signed, err := types.SignTx(tx, types.NewLondonSigner(s.chainID), s.gasPayer)
	if err != nil {
		return "", err
	}

	if err := s.cli.SendTransaction(ctx, signed); err != nil {
		return "", fmt.Errorf("send tx: %w", err)
	}

	receipt, err := bind.WaitMined(ctx, s.cli, signed)
	if err != nil {
		return "", fmt.Errorf("wait mined: %w", err)
	}

	if receipt.Status == types.ReceiptStatusFailed {
		return "", fmt.Errorf("tx reverted on-chain")
	}

	return signed.Hash().Hex(), nil
}
