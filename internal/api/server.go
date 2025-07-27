package api

import (
	"crypto/ecdsa"
	"fmt"
	"net/http"
	"os"

	auth "x402/internal/authorizers/exact"
	"x402/internal/config"
	settler "x402/internal/settlers/exact"
	"x402/pkg/x402"

	"github.com/alexedwards/flow"
	"github.com/ethereum/go-ethereum/accounts/keystore"
)

type Server struct {
	router      *flow.Mux
	authorizers map[string]x402.Authorizer
	settlers    map[string]x402.Settler
	cfg         config.Config
}

func NewServer(cfg config.Config) (*Server, error) {
	s := &Server{
		router:      flow.New(),
		authorizers: make(map[string]x402.Authorizer),
		settlers:    make(map[string]x402.Settler),
		cfg:         cfg,
	}

	err := s.initAuthorizers()
	if err != nil {
		return nil, err
	}

	err = s.initSettlers()
	if err != nil {
		return nil, err
	}

	s.setupRoutes()

	return s, nil
}

func key(scheme, network string) string { return scheme + "|" + network }

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *Server) setupRoutes() {
	s.router.HandleFunc("/verify", s.handleVerify, http.MethodPost)
	s.router.HandleFunc("/settle", s.handleSettle, http.MethodPost)
	s.router.HandleFunc("/supported", s.handleSupported, http.MethodGet)
}

func (s *Server) initAuthorizers() error {
	usdcExactAuth, err := auth.NewEIP3009Authorizer(s.cfg.RPCURL, s.cfg.USDCAddress)
	if err != nil {
		return err
	}

	s.authorizers[key("exact", "ethereum")] = usdcExactAuth

	return nil
}

func (s *Server) initSettlers() error {
	var gasKey *ecdsa.PrivateKey
	switch {
	case s.cfg.KeyStoreFile != "":
		data, err := os.ReadFile(s.cfg.KeyStoreFile)
		if err != nil {
			return fmt.Errorf("read keystore: %w", err)
		}
		key, err := keystore.DecryptKey(data, s.cfg.KeystorePass)
		if err != nil {
			return fmt.Errorf("decrypt keystore: %w", err)
		}
		gasKey = key.PrivateKey
	default:
		return fmt.Errorf("no gas key configured (provide KEYSTORE_FILE/KEYSTORE_PASS or GAS_PRIVATE_KEY)")
	}

	exactSettler, err := settler.NewEIP3009SettlerCore(s.cfg.RPCURL, s.cfg.USDCAddress, gasKey)
	if err != nil {
		return err
	}
	s.settlers[key("exact", "ethereum")] = exactSettler
	return nil
}
