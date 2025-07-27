package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"x402/internal/config"
	"x402/pkg/x402"
)

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
	pay := x402.PaymentPayload{
		X402Version: 1,
		Scheme:      "exact",
		Network:     "ethereum",
		Payload: &x402.ExactEvmPayload{
			Signature: "0x1",
			Authorization: &x402.ExactEvmPayloadAuthorization{
				From:        "0x1111111111111111111111111111111111111111",
				To:          "0x1111111111111111111111111111111111111111",
				Value:       "100000",
				ValidAfter:  "0",
				ValidBefore: "4102444800",
				Nonce:       "deadbeefcafebabe000000000000000000000000000000000000000000000000",
			},
		},
	}

	payloadJSON, err := json.Marshal(pay)
	if err != nil {
		t.Fatalf("FATAL ERROR: could not marshall struct to json: %v", err)
	}

	req := x402.VerifyRequest{
		X402Version:   1,
		PaymentHeader: base64.StdEncoding.EncodeToString(payloadJSON),
		PaymentRequirements: x402.PaymentRequirements{
			Scheme:            "exact",
			Network:           "ethereum",
			PayTo:             "0x1111111111111111111111111111111111111111",
			MaxAmountRequired: "100000",
		},
	}

	vreq, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("error marshalling Verify Request struct into JSON, err: %v", err)
	}

	srv, err := newTestServer()
	if err != nil {
		t.Fatalf("error creating server: %v", err)
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

func TestConfig(t *testing.T) {
	testConfig, err := config.Load("../config/config.env")
	if err != nil {
		t.Fatalf("error loading config: %v", err)
	}

	t.Log(testConfig)

	url, err := config.GrabRPCUrl("../config/config.env")
	if err != nil {
		t.Fatalf("error grabbing url: %v", err)
	}

	t.Logf("URL: %s", url)
}
