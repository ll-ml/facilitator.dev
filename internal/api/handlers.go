package api

import (
	"context"
	"encoding/json"
	"net/http"

	"x402/pkg/x402"
)

// POST /verify
func (s *Server) handleVerify(w http.ResponseWriter, r *http.Request) {
	if ct := r.Header.Get("Content-Type"); ct != "application/json" {
		http.Error(w, "Content Type must be application/json", http.StatusUnsupportedMediaType) // This should be JSON at some point
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // max 1 MiB JSON body
	defer r.Body.Close()

	var req x402.VerifyRequest
	secureRequestDecoder := json.NewDecoder(r.Body)
	secureRequestDecoder.DisallowUnknownFields()
	if err := secureRequestDecoder.Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	paymentPayload, err := x402.DecodePaymentPayloadFromBase64(req.PaymentHeader)
	if err != nil {
		resp := x402.NewVerifyFail("Invalid payment header encoding")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	if paymentPayload.Scheme != req.PaymentRequirements.Scheme {
		resp := x402.NewVerifyFail("Payment scheme does not match requirements")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	if paymentPayload.Network != req.PaymentRequirements.Network {
		resp := x402.NewVerifyFail("Payment network does not match requirements")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	au, ok := s.authorizers[key(req.PaymentRequirements.Scheme, req.PaymentRequirements.Network)]
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode((x402.NewVerifyFail("unsupported scheme/network")))
	}

	ctx, cancel := context.WithTimeout(r.Context(), s.cfg.VerifyTimeout)
	defer cancel()

	res := au.Authorize(ctx, paymentPayload, &req.PaymentRequirements)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (s *Server) handleSettle(w http.ResponseWriter, r *http.Request) {
	if ct := r.Header.Get("Content-Type"); ct != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	defer r.Body.Close()

	var req x402.SettleRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusUnsupportedMediaType)
		return
	}

	payload, err := x402.DecodePaymentPayloadFromBase64(req.PaymentHeader)
	if err != nil {
		http.Error(w, "invalid payment header format", http.StatusBadRequest) // if we cant decode it assume its client error
		return
	}

	settler, ok := s.settlers[key(req.PaymentRequirements.Scheme, req.PaymentRequirements.Network)]
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(x402.SettleResponse{
			Success:     false,
			ErrorReason: strPtr("unsupported scheme/network"),
		})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), s.cfg.SettleTimeout)
	defer cancel()

	txHash, err := settler.Settle(ctx, payload, &req.PaymentRequirements)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(x402.SettleResponse{
			Success:     false,
			ErrorReason: strPtr(err.Error()),
		})
		return
	}

	json.NewEncoder(w).Encode(x402.SettleResponse{
		Success:     true,
		Transaction: txHash,
		Network:     req.PaymentRequirements.Network, // Make sure to include payer in this later on
	})
}

// GET /supported - Exactly as specified
func (s *Server) handleSupported(w http.ResponseWriter, _ *http.Request) {
	response := x402.SupportedResponse{
		Kinds: []struct {
			Scheme  string `json:"scheme"`
			Network string `json:"network"`
		}{
			{Scheme: "exact", Network: "ethereum"},
			//{Scheme: "exact", Network: "solana"}, only usdc for now for MVP
			//{Scheme: "exact", Network: "polygon"},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func strPtr(s string) *string {
	return &s
}
