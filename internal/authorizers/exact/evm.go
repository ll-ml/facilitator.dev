package exact

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"x402/internal/cryptohelpers"
	"x402/pkg/x402"
)

var (
	// keccak256("EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)")
	eip712DomainTypeHash = crypto.Keccak256Hash([]byte(
		"EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)",
	))

	// keccak256("TransferWithAuthorization(address from,address to,uint256 value,uint256 validAfter,uint256 validBefore,bytes32 nonce)")
	transferWithAuthTypeHash = crypto.Keccak256Hash([]byte(
		"TransferWithAuthorization(address from,address to,uint256 value,uint256 validAfter,uint256 validBefore,bytes32 nonce)",
	))
)

type evmAuthorizer struct {
	rpc           ethereum.ContractCaller
	abi           abi.ABI
	usdc          common.Address
	chainID       *big.Int
	verifyTimeout time.Duration
	domain        EIP3009Domain
}

type EIP3009Domain struct {
	Name    string
	Version string
	ChainID *big.Int
	Token   common.Address
}

func NewEIP3009Authorizer(rpcURL string, usdcAddr string) (x402.Authorizer, error) {
	rpc, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, err
	}

	return NewEIP3009AuthorizerCore(rpc, common.HexToAddress(usdcAddr))
}

func NewEIP3009AuthorizerCore(rpc ethereum.ContractCaller, usdc common.Address) (x402.Authorizer, error) {
	parsed, err := abi.JSON(strings.NewReader(usdcEIP3009ABI))
	if err != nil {
		return nil, err
	}

	chainID := big.NewInt(1)

	dom := EIP3009Domain{
		Name:    "USD Coin",
		Version: "2",
		ChainID: chainID,
		Token:   usdc,
	}

	return &evmAuthorizer{
		rpc:     rpc,
		abi:     parsed,
		usdc:    usdc,
		chainID: chainID,
		domain:  dom,
	}, nil
}

func (e *evmAuthorizer) Authorize(
	ctx context.Context,
	p *x402.PaymentPayload,
	req *x402.PaymentRequirements,
) x402.VerifyResponse {
	if p.Scheme != "exact" || p.Network != "ethereum" {
		return x402.NewVerifyFail("unsupported scheme/network")
	}

	var evm x402.ExactEvmPayload
	raw, _ := json.Marshal(p.Payload)
	if err := json.Unmarshal(raw, &evm); err != nil {
		return x402.NewVerifyFail("malformed scheme payload")
	}

	auth := evm.Authorization
	if auth == nil {
		return x402.NewVerifyFail("authorization missing")
	}

	if !strings.EqualFold(auth.To, req.PayTo) {
		return x402.NewVerifyFail("recipient mismatch")
	}

	authAmt, ok := new(big.Int).SetString(strings.TrimSpace(auth.Value), 10)
	if !ok {
		return x402.NewVerifyFail("invalid auth value")
	}

	maxAmt, ok := new(big.Int).SetString(strings.TrimSpace(req.MaxAmountRequired), 10)
	if !ok {
		return x402.NewVerifyFail("invalid maxAmountRequired")
	}

	if authAmt.Cmp(maxAmt) != 0 {
		return x402.NewVerifyFail("amount mismatch")
	}

	now := big.NewInt(time.Now().Unix())
	validAfter, _ := new(big.Int).SetString(auth.ValidAfter, 10)
	validBefore, _ := new(big.Int).SetString(auth.ValidBefore, 10)
	if now.Cmp(validAfter) < 0 || now.Cmp(validBefore) > 0 {
		return x402.NewVerifyFail("authorization not currently valid")
	}

	ok, recovered, err := VerifyEIP3009Sig(e.domain, evm)
	if !ok {
		msg := "invalid signature"
		if err != nil {
			msg += ": " + err.Error()
		}
		return x402.NewVerifyFail(msg)
	}

	if !strings.EqualFold(recovered.Hex(), auth.From) {
		return x402.NewVerifyFail("recovered signer != from")
	}

	v, r, s, err := cryptohelpers.SplitSig(evm.Signature)
	if err != nil {
		return x402.NewVerifyFail("bad signature encoding: " + err.Error())
	}

	nonce32, err := cryptohelpers.HexTo32(auth.Nonce)
	if err != nil {
		return x402.NewVerifyFail("bad nonce: " + err.Error())
	}

	from := common.HexToAddress(auth.From)
	to := common.HexToAddress(auth.To)

	// eth_call simulation to ensure it won't revert (no gas cost)
	calldata, err := e.abi.Pack(
		"transferWithAuthorization",
		from,
		to,
		authAmt,
		validAfter,
		validBefore,
		nonce32,
		v, r, s,
	)
	if err != nil {
		return x402.NewVerifyFail("abi pack: " + err.Error())
	}
	callMsg := ethereum.CallMsg{To: &e.usdc, Data: calldata}
	cctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if _, err := e.rpc.CallContract(cctx, callMsg, nil); err != nil {
		return x402.NewVerifyFail("simulation revert: " + err.Error())
	}

	return x402.NewVerifyOK(auth.From)
}

// VerifyEIP3009Sig verifies an EIP-3009 TransferWithAuthorization signature off-chain.
// It returns (ok, recoveredSigner, error).
func VerifyEIP3009Sig(
	d EIP3009Domain,
	evm x402.ExactEvmPayload,
) (bool, common.Address, error) {

	if evm.Authorization == nil {
		return false, common.Address{}, errors.New("missing authorization")
	}
	auth := evm.Authorization

	// 1) decode hex -> 65 bytes
	sigBytes, err := hexutil.Decode(evm.Signature)
	if err != nil {
		return false, common.Address{}, fmt.Errorf("bad signature hex: %w", err)
	}
	if len(sigBytes) != 65 {
		return false, common.Address{}, fmt.Errorf("bad signature length %d", len(sigBytes))
	}

	// 2) split r,s,v
	r := new(big.Int).SetBytes(sigBytes[0:32])
	s := new(big.Int).SetBytes(sigBytes[32:64])
	vRaw := sigBytes[64]

	// 3) normalize v to {27,28}
	var v27 byte
	switch vRaw {
	case 27, 28:
		v27 = vRaw
	case 0, 1:
		v27 = vRaw + 27
	default:
		return false, common.Address{}, fmt.Errorf("invalid v %d", vRaw)
	}

	if err := validateSigValues(v27, r, s); err != nil {
		return false, common.Address{}, fmt.Errorf(
			"invalid signature values: vRaw=%d normalized=%d r=%s s=%s (%v)",
			vRaw, v27, r.Text(16), s.Text(16), err,
		)
	}

	// 5) make a COPY and downshift v to 0/1 for SigToPub
	sigForRecover := make([]byte, 65)
	copy(sigForRecover, sigBytes)
	sigForRecover[64] = v27 - 27 // now 0/1

	// 6) domain + struct hash
	domainSep := hashDomain(d)
	structHash, err := hashTransferWithAuth(auth)
	if err != nil {
		return false, common.Address{}, err
	}
	digest := eip191Digest(domainSep, structHash)

	// 7) recover
	pub, err := crypto.SigToPub(digest.Bytes(), sigForRecover)
	if err != nil {
		return false, common.Address{}, fmt.Errorf("ecrecover failed: %w", err)
	}
	recovered := crypto.PubkeyToAddress(*pub)

	// 8) compare to "from"
	from := common.HexToAddress(auth.From)
	if !bytes.Equal(recovered.Bytes(), from.Bytes()) {
		return false, recovered, errors.New("recovered signer != from")
	}
	return true, recovered, nil
}

func validateSigValues(v byte, r, s *big.Int) error {
	if v != 27 && v != 28 {
		return fmt.Errorf("bad v %d", v)
	}
	if r.Sign() <= 0 || s.Sign() <= 0 {
		return fmt.Errorf("r or s is zero/negative")
	}
	n := crypto.S256().Params().N
	if r.Cmp(n) >= 0 {
		return fmt.Errorf("r >= curve order")
	}
	if s.Cmp(n) >= 0 {
		return fmt.Errorf("s >= curve order")
	}
	halfN := new(big.Int).Rsh(new(big.Int).Set(n), 1)
	if s.Cmp(halfN) > 0 {
		return fmt.Errorf("s is not low-S")
	}
	return nil
}

func hashDomain(d EIP3009Domain) common.Hash {
	nameHash := crypto.Keccak256([]byte(d.Name))
	versionHash := crypto.Keccak256([]byte(d.Version))

	var b bytes.Buffer
	b.Write(eip712DomainTypeHash.Bytes())
	b.Write(nameHash)
	b.Write(versionHash)
	b.Write(padBig(d.ChainID))
	b.Write(padAddress(d.Token))

	return crypto.Keccak256Hash(b.Bytes())
}

func hashTransferWithAuth(a *x402.ExactEvmPayloadAuthorization) (common.Hash, error) {
	from := common.HexToAddress(a.From)
	to := common.HexToAddress(a.To)

	val, ok := new(big.Int).SetString(a.Value, 10)
	if !ok {
		return common.Hash{}, fmt.Errorf("bad value: %s", a.Value)
	}
	validAfter, ok := new(big.Int).SetString(a.ValidAfter, 10)
	if !ok {
		return common.Hash{}, fmt.Errorf("bad validAfter: %s", a.ValidAfter)
	}
	validBefore, ok := new(big.Int).SetString(a.ValidBefore, 10)
	if !ok {
		return common.Hash{}, fmt.Errorf("bad validBefore: %s", a.ValidBefore)
	}
	nonceBytes, err := cryptohelpers.HexTo32(a.Nonce)
	if err != nil {
		return common.Hash{}, fmt.Errorf("bad nonce: %w", err)
	}

	var b bytes.Buffer
	b.Write(transferWithAuthTypeHash.Bytes())
	b.Write(padAddress(from))
	b.Write(padAddress(to))
	b.Write(padBig(val))
	b.Write(padBig(validAfter))
	b.Write(padBig(validBefore))
	b.Write(nonceBytes[:])

	return crypto.Keccak256Hash(b.Bytes()), nil
}

func padBig(n *big.Int) []byte {
	out := make([]byte, 32)
	if n == nil {
		return out
	}
	b := n.Bytes()
	copy(out[32-len(b):], b)
	return out
}

func eip191Digest(domain, data common.Hash) common.Hash {
	prefix := []byte{0x19, 0x01}
	out := crypto.Keccak256Hash(
		append(append(prefix, domain.Bytes()...), data.Bytes()...),
	)
	return out
}

func padAddress(a common.Address) []byte {
	out := make([]byte, 32)
	copy(out[12:], a.Bytes()) // right-most 20 bytes
	return out
}
