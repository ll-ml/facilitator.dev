package cryptohelpers

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

func HexTo32(h string) ([32]byte, error) {
	var out [32]byte
	h = strings.TrimPrefix(h, "0x")
	b, err := hex.DecodeString(h)
	if err != nil {
		return out, err
	}
	if len(b) > 32 {
		return out, fmt.Errorf("nonce too long")
	}
	copy(out[32-len(b):], b)
	return out, nil
}

func SplitSig(sigHex string) (v uint8, r [32]byte, s [32]byte, err error) {
	b, err := hexutil.Decode(sigHex)
	if err != nil {
		return 0, r, s, err
	}
	if len(b) != 65 {
		return 0, r, s, fmt.Errorf("len %d != 65", len(b))
	}
	copy(r[:], b[0:32])
	copy(s[:], b[32:64])
	v = b[64]
	if v < 27 {
		v += 27
	}
	return v, r, s, nil
}
