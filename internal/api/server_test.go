package api

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"x402/internal/config"
	"x402/internal/cryptohelpers"
	"x402/pkg/x402"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

const usdcSepola = "0x1c7D4B196Cb0C7B01d743Fbc6116a902379C7238"

func newTestServer() (*Server, error) {
	srvCfg, err := config.Load("./config.env")
	if err != nil {
		return nil, err
	}

	srv, err := NewServer(srvCfg)
	if err != nil {
		return nil, err
	}

	return srv, nil
}

func TestSupportedEndpoint(t *testing.T) {
	srv, err := newTestServer()
	if err != nil {
		t.Fatalf("error creating server: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/supported", nil)
	rr := httptest.NewRecorder()

	srv.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	supportedResonse := new(x402.SupportedResponse)
	if err := json.NewDecoder(rr.Body).Decode(&supportedResonse); err != nil {
		t.Fatalf("failed to decode response in correct struct err: %v", err)
	}

}

func TestAgainstLargeReqBody(t *testing.T) {
	srv, err := newTestServer()
	if err != nil {
		t.Fatalf("error creating server: %v", err)
	}

	largeReqBody := strings.Repeat("this is my request", 10<<20) // 10 MiB worth of junk

	badRequest := httptest.NewRequest(http.MethodPost, "/verify", strings.NewReader(largeReqBody))
	rr := httptest.NewRecorder()

	srv.ServeHTTP(rr, badRequest)

	if rr.Code < 400 {
		t.Fatalf("expected error status code, got %d", rr.Code)
	}

}

func TestVerifyEndPoint(t *testing.T) {
	tempKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}

	sellerKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}

	from := crypto.PubkeyToAddress(tempKey.PublicKey)
	to := crypto.PubkeyToAddress(sellerKey.PublicKey)

	val := big.NewInt(0)

	auth := &x402.ExactEvmPayloadAuthorization{
		From:        from.Hex(),
		To:          to.Hex(),
		Value:       val.String(),
		ValidAfter:  "0",
		ValidBefore: "4102444800",
		Nonce:       "0x" + strings.Repeat("ab", 32),
	}

	srv, err := newTestServer()
	if err != nil {
		t.Fatalf("error creating server: %v", err)
	}

	dom := cryptohelpers.EIP3009Domain{
		Name:    "USDC",
		Version: "2",
		ChainID: big.NewInt(11155111), // this should be 1337 when doing actual settle
		Token:   common.HexToAddress(srv.cfg.USDCAddress),
	}

	sig, err := SignEIP3009Authorization(tempKey, dom, auth)
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

	payloadJSON, err := json.Marshal(pay)
	if err != nil {
		t.Fatalf("FATAL ERROR: could not marshall struct to json: %v", err)
	}

	paymentReq := x402.PaymentRequirements{
		Scheme:            "exact",
		Network:           "ethereum",
		PayTo:             auth.To,
		MaxAmountRequired: auth.Value,
		Asset:             srv.cfg.USDCAddress,
		Resource:          "test://unit",
		Description:       "unit test transferWithAuthorization",
		MimeType:          "application/json",
		MaxTimeoutSeconds: 600,
	}

	req := x402.VerifyRequest{
		X402Version:         1,
		PaymentHeader:       base64.StdEncoding.EncodeToString(payloadJSON),
		PaymentRequirements: paymentReq,
	}

	vreq, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("error marshalling Verify Request struct into JSON, err: %v", err)
	}

	rr := httptest.NewRecorder()
	simReq := httptest.NewRequest(http.MethodPost, "/verify", bytes.NewReader(vreq))
	simReq.Header.Set("Content-Type", "application/json")

	srv.ServeHTTP(rr, simReq)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rr.Code)
	}

	serversResponse := new(x402.VerifyResponse)
	if err := json.NewDecoder(rr.Body).Decode(serversResponse); err != nil {
		t.Fatalf("error marshalling JSON from server response: %v", err)
	}

	if !serversResponse.IsValid {
		t.Logf("Blockchain hit however sim reverted with message: %s", *serversResponse.InvalidReason) // This will be non fatal as long as verify is dummy
	}

}

func TestSettle_ExactEVM_Sepola(t *testing.T) {
	buyerKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	from := crypto.PubkeyToAddress(buyerKey.PublicKey)

	sellerKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	to := crypto.PubkeyToAddress(sellerKey.PublicKey)

	srv, err := newTestServer()
	if err != nil {
		t.Fatal(err)
	}

	auth := &x402.ExactEvmPayloadAuthorization{
		From:        from.Hex(),
		To:          to.Hex(),
		Value:       "0",
		ValidAfter:  "0",
		ValidBefore: "4102444800",
		Nonce:       "0x" + strings.Repeat("ab", 32),
	}

	// EIP‑712 domain for **Sepolia USDC**
	dom := cryptohelpers.EIP3009Domain{
		Name:    "USDC",
		Version: "2",
		ChainID: big.NewInt(11155111),
		Token:   common.HexToAddress(srv.cfg.USDCAddress),
	}

	sig, err := SignEIP3009Authorization(buyerKey, dom, auth)
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

	// Optional local sanity check
	if ok, _, err := VerifyEIP3009Sig(dom, *pay.Payload); !ok {
		t.Fatalf("local verify failed: %v", err)
	}

	// Build PaymentRequirements (must match payload)
	reqs := x402.PaymentRequirements{
		Scheme:            "exact",
		Network:           "ethereum",
		PayTo:             auth.To,
		MaxAmountRequired: auth.Value,          // "0"
		Asset:             srv.cfg.USDCAddress, // Sepolia USDC addr
		Resource:          "test://settle",
		Description:       "e2e settle test",
		MimeType:          "application/json",
		MaxTimeoutSeconds: 600,
	}

	// 1) /verify (eth_call) should pass
	hdrBytes, _ := json.Marshal(pay)
	verifyBody, _ := json.Marshal(x402.VerifyRequest{
		X402Version:         1,
		PaymentHeader:       base64.StdEncoding.EncodeToString(hdrBytes),
		PaymentRequirements: reqs,
	})
	rr := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/verify", bytes.NewReader(verifyBody))
	r.Header.Set("Content-Type", "application/json")
	srv.ServeHTTP(rr, r)

	if rr.Code != http.StatusOK {
		t.Fatalf("/verify status=%d body=%s", rr.Code, rr.Body.String())
	}
	var vres x402.VerifyResponse
	_ = json.NewDecoder(rr.Body).Decode(&vres)
	if !vres.IsValid {
		t.Fatalf("/verify invalid: %s", *vres.InvalidReason)
	}

	// 2) /settle should send a real tx (gas payer pays)
	settleBody, _ := json.Marshal(x402.SettleRequest{
		X402Version:         1,
		PaymentHeader:       base64.StdEncoding.EncodeToString(hdrBytes),
		PaymentRequirements: reqs,
	})

	rr2 := httptest.NewRecorder()
	r2 := httptest.NewRequest(http.MethodPost, "/settle", bytes.NewReader(settleBody))
	r2.Header.Set("Content-Type", "application/json")
	srv.ServeHTTP(rr2, r2)

	if rr2.Code != http.StatusOK {
		t.Fatalf("/settle status=%d body=%s", rr2.Code, rr2.Body.String())
	}
	var sres x402.SettleResponse
	_ = json.NewDecoder(rr2.Body).Decode(&sres)
	if !sres.Success || sres.Transaction == "" {
		t.Fatalf("/settle failed: %v", *sres.ErrorReason)
	}
	t.Logf("settled tx: %s on %s", sres.Transaction, sres.Network)
}

func TestMaliciousReqHeaders(t *testing.T) {
	srv, err := newTestServer()
	if err != nil {
		t.Fatalf("error creating server: %v", err)
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/verify", bytes.NewReader([]byte("Hello world")))
	req.Header.Set("Content-Type", "text/plain")

	srv.ServeHTTP(rr, req)

	if rr.Code != 415 {
		t.Fatalf("error, expected to get status 415 with malicious request. Got: %d", rr.Code)
	}
}

// Need to move this code at some point

func SignEIP3009Authorization(priv *ecdsa.PrivateKey, dom cryptohelpers.EIP3009Domain, auth *x402.ExactEvmPayloadAuthorization) (string, error) {
	domainSep := cryptohelpers.HashDomain(dom)

	if _, ok := new(big.Int).SetString(auth.Value, 10); !ok {
		return "", fmt.Errorf("bad value")
	}

	if _, ok := new(big.Int).SetString(auth.ValidAfter, 10); !ok {
		return "", fmt.Errorf("bad validAfter")
	}

	if _, ok := new(big.Int).SetString(auth.ValidBefore, 10); !ok {
		return "", fmt.Errorf("bad validBefore")
	}

	structHash, err := cryptohelpers.HashTransferWithAuth(auth)
	if err != nil {
		return "", err
	}
	digest := cryptohelpers.Eip191Digest(domainSep, structHash)

	sig, err := crypto.Sign(digest.Bytes(), priv)
	if err != nil {
		return "", err
	}

	sig[64] += 27

	return hexutil.Encode(sig), nil
}

// VerifyEIP3009Sig verifies an EIP-3009 TransferWithAuthorization signature off-chain.
// It returns (ok, recoveredSigner, error).
func VerifyEIP3009Sig(
	d cryptohelpers.EIP3009Domain,
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
	domainSep := cryptohelpers.HashDomain(d)
	structHash, err := cryptohelpers.HashTransferWithAuth(auth)
	if err != nil {
		return false, common.Address{}, err
	}
	digest := cryptohelpers.Eip191Digest(domainSep, structHash)

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
