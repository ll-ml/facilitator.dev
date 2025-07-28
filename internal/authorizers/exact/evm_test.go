package exact

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"
	"testing"
	"time"
	"x402/internal/config"
	"x402/pkg/x402"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

const usdcMainnet = "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"
const usdcSepola = "0x1c7D4B196Cb0C7B01d743Fbc6116a902379C7238"

func newSimBackend() (*backends.SimulatedBackend, common.Address) {
	key, _ := crypto.GenerateKey()
	from := crypto.PubkeyToAddress(key.PublicKey)

	alloc := types.GenesisAlloc{from: {Balance: big.NewInt(1e18)}}
	sim := backends.NewSimulatedBackend(alloc, 15_000_000)

	auth, _ := bind.NewKeyedTransactorWithChainID(key, big.NewInt(1337))
	usdc, _, _, _ := DeployExact(auth, sim)
	sim.Commit()
	return sim, usdc
}

func buildDummyPayload(from, to common.Address, value *big.Int) *x402.PaymentPayload {
	return &x402.PaymentPayload{
		X402Version: 1,
		Scheme:      "exact",
		Network:     "ethereum",
		Payload: &x402.ExactEvmPayload{
			Signature: "0x12",
			Authorization: &x402.ExactEvmPayloadAuthorization{
				From:        from.Hex(),
				To:          to.Hex(),
				Value:       value.String(),
				ValidAfter:  "0",
				ValidBefore: "4102444800",
				Nonce:       "0xabc",
			},
		},
	}
}

func buildDummyReq(to common.Address, value *big.Int) x402.PaymentRequirements {
	return x402.PaymentRequirements{
		Scheme:            "exact",
		Network:           "ethereum",
		PayTo:             to.Hex(),
		MaxAmountRequired: value.String(),
	}
}

func SignEIP3009Authorization(priv *ecdsa.PrivateKey, dom EIP3009Domain, auth *x402.ExactEvmPayloadAuthorization) (string, error) {
	domainSep := hashDomain(dom)

	if _, ok := new(big.Int).SetString(auth.Value, 10); !ok {
		return "", fmt.Errorf("bad value")
	}

	if _, ok := new(big.Int).SetString(auth.ValidAfter, 10); !ok {
		return "", fmt.Errorf("bad validAfter")
	}

	if _, ok := new(big.Int).SetString(auth.ValidBefore, 10); !ok {
		return "", fmt.Errorf("bad validBefore")
	}

	structHash, err := hashTransferWithAuth(auth)
	if err != nil {
		return "", err
	}
	digest := eip191Digest(domainSep, structHash)

	sig, err := crypto.Sign(digest.Bytes(), priv)
	if err != nil {
		return "", err
	}

	sig[64] += 27

	return hexutil.Encode(sig), nil
}

func TestAuthorize_ExactEVM_Infura(t *testing.T) {
	rpcUrl, err := config.GrabRPCUrl("../../config/config.env")
	if err != nil {
		t.Fatalf("error loading config file: %v", err)
	}

	au, err := NewEIP3009Authorizer(rpcUrl, usdcSepola)
	if err != nil {
		t.Fatalf("error creating ACTUAl authorizer: %v", err)
	}

	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("error generating key: %v", err)
	}
	from := crypto.PubkeyToAddress(key.PublicKey)

	key2, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("error generating key %v", err)
	}

	to := crypto.PubkeyToAddress(key2.PublicKey)
	val := big.NewInt(1_000_000)

	pay := buildDummyPayload(from, to, val)
	req := x402.PaymentRequirements{
		Scheme:            "exact",
		Network:           "ethereum",
		PayTo:             to.Hex(),
		MaxAmountRequired: val.String(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	res := au.Authorize(ctx, pay, &req)
	if !res.IsValid {
		t.Fatalf("eth_call failed: %v", *res.InvalidReason)
	}

	t.Logf("Infura simulation succeeded, payer=%s", *res.Payer)
}

func TestAuthorize_ExactEVM_Simulated(t *testing.T) {
	sim, usdc := newSimBackend()

	au, err := NewEIP3009AuthorizerCore(sim, usdc)
	if err != nil {
		t.Fatalf("error could not create authorizer: %v", err)
	}

	from := common.HexToAddress("0x1111111111111111111111111111111111111111")
	to := common.HexToAddress("0x2222222222222222222222222222222222222222")
	val := big.NewInt(1_000_000)
	pay := buildDummyPayload(from, to, val)
	req := buildDummyReq(to, val)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res := au.Authorize(ctx, pay, &req)
	if !res.IsValid {
		t.Fatalf("expected valid response got: %v", *res.InvalidReason)
	}
	t.Logf("RESPONSE OBJ: %v", res)
}

func TestPrintHash(t *testing.T) {
	key, _ := crypto.GenerateKey()
	from := crypto.PubkeyToAddress(key.PublicKey)

	key2, _ := crypto.GenerateKey()
	to := crypto.PubkeyToAddress(key2.PublicKey)

	val := big.NewInt(0)
	auth := &x402.ExactEvmPayloadAuthorization{
		From:        from.Hex(),
		To:          to.Hex(),
		Value:       val.String(),
		ValidAfter:  "0",
		ValidBefore: "4102444800",
		Nonce:       "0x" + strings.Repeat("ab", 32),
	}

	dom := EIP3009Domain{
		Name:    "USDC",
		Version: "2",
		ChainID: big.NewInt(11155111), // this should be 1337 when doing actual settle
		Token:   common.HexToAddress(usdcSepola),
	}

	sig, err := SignEIP3009Authorization(key, dom, auth)
	if err != nil {
		t.Fatal(err)
	}

	pay := &x402.PaymentPayload{
		X402Version: 1,
		Scheme:      "exact",
		Network:     "ethereum",
		Payload: &x402.ExactEvmPayload{
			Signature:     sig,
			Authorization: auth,
		},
	}

	ok, _, err := VerifyEIP3009Sig(dom, *pay.Payload)
	if !ok {
		t.Fatal(err)
	}

	req := &x402.PaymentRequirements{
		Scheme:            "exact",
		Network:           "ethereum",
		PayTo:             auth.To,
		MaxAmountRequired: auth.Value,
		Asset:             usdcSepola,
		Resource:          "test://unit",
		Description:       "unit test transferWithAuthorization",
		MimeType:          "application/json",
		MaxTimeoutSeconds: 600,
	}

	rpcUrl, err := config.GrabRPCUrl("../../config/config.env")
	if err != nil {
		t.Fatalf("error loading config file: %v", err)
	}
	t.Log(rpcUrl)

	au, err := NewEIP3009Authorizer(rpcUrl, usdcSepola)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	res := au.Authorize(ctx, pay, req)
	if !res.IsValid {
		t.Fatalf("authorization failed: %s", *res.InvalidReason)
	}

	t.Log("SUCCES!!!")
	t.Log(res.IsValid, res.Payer)
}
