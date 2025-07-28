package cryptohelpers

import (
	"bytes"
	"fmt"
	"math/big"
	"x402/pkg/x402"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type EIP3009Domain struct {
	Name    string
	Version string
	ChainID *big.Int
	Token   common.Address
}

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

func HashTransferWithAuth(a *x402.ExactEvmPayloadAuthorization) (common.Hash, error) {
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
	nonceBytes, err := HexTo32(a.Nonce)
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

func HashDomain(d EIP3009Domain) common.Hash {
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

func padAddress(a common.Address) []byte {
	out := make([]byte, 32)
	copy(out[12:], a.Bytes()) // right-most 20 bytes
	return out
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

func Eip191Digest(domain, data common.Hash) common.Hash {
	prefix := []byte{0x19, 0x01}
	out := crypto.Keccak256Hash(
		append(append(prefix, domain.Bytes()...), data.Bytes()...),
	)
	return out
}
