package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

type Config struct {
	HTTPAddr      string
	RPCURL        string // will be fully assembled URL with key
	USDCAddress   string // addr of usdc on whatever network your using
	VerifyTimeout time.Duration
	SettleTimeout time.Duration

	KeyStoreFile string
	KeystorePass string

	GasPrivatekey string
}

func Load(path string) (Config, error) {
	envMap, err := parseEnvFile(path)
	if err != nil {
		return Config{}, err
	}

	rpcBase, ok := envMap["RPC"]
	if !ok || strings.TrimSpace(rpcBase) == "" {
		return Config{}, fmt.Errorf("missing RPC in %s", path)
	}
	apiKey, ok := envMap["KEY"]
	if !ok || strings.TrimSpace(apiKey) == "" {
		return Config{}, fmt.Errorf("missing KEY in %s", path)
	}

	// Build the URL safely. Infura pattern is "<base>/v3/<key>" usually.
	rpcURL := strings.TrimRight(rpcBase, "/")
	if !strings.HasSuffix(rpcURL, "/v3") && !strings.Contains(rpcURL, apiKey) {
		rpcURL = fmt.Sprintf("%s/v3/%s", rpcURL, apiKey)
	} else if !strings.Contains(rpcURL, apiKey) {
		rpcURL = fmt.Sprintf("%s/%s", rpcURL, apiKey)
	}

	cfg := Config{
		HTTPAddr:      getOr(envMap, "HTTP_ADDR", ":8080"),
		RPCURL:        rpcURL,
		USDCAddress:   getOr(envMap, "USDC_ADDRESS", "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"),
		VerifyTimeout: getDuration(envMap, "VERIFY_TIMEOUT", 3*time.Second),
		SettleTimeout: getDuration(envMap, "SETTLE_TIMEOUT", 10*time.Second),

		KeyStoreFile:  getOr(envMap, "KEYSTORE_FILE", ""),
		KeystorePass:  getOr(envMap, "KEYSTORE_PASS", ""),
		GasPrivatekey: getOr(envMap, "GAS_PRIVATE_KEY", ""),
	}
	return cfg, nil
}

func GrabRPCUrl(path string) (string, error) {
	envMap, err := parseEnvFile(path)
	if err != nil {
		return "", err
	}

	rpcBase, ok := envMap["RPC"]
	if !ok || strings.TrimSpace(rpcBase) == "" {
		return "", fmt.Errorf("missing RPC in %s", path)
	}

	apiKey, ok := envMap["KEY"]
	if !ok || strings.TrimSpace(apiKey) == "" {
		return "", fmt.Errorf("missing KEY in %s", path)
	}

	rpcUrl := strings.TrimRight(rpcBase, "/")
	if !strings.HasSuffix(rpcUrl, "/v3") && !strings.Contains(rpcUrl, apiKey) {
		rpcUrl = fmt.Sprintf("%s/v3/%s", rpcUrl, apiKey)
	} else if !strings.Contains(rpcUrl, apiKey) {
		rpcUrl = fmt.Sprintf("%s/%s", rpcUrl, apiKey)
	}

	return rpcUrl, err
}

func parseEnvFile(envFilePath string) (map[string]string, error) {
	envMap := make(map[string]string)

	envFile, err := os.Open(envFilePath)
	if err != nil {
		return nil, err
	}
	defer envFile.Close()

	envFileScanner := bufio.NewScanner(envFile)
	for envFileScanner.Scan() {
		line := strings.TrimSpace(envFileScanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		tokens := strings.SplitN(line, "=", 2)
		if len(tokens) != 2 {
			return nil, fmt.Errorf("error: malformed line in env file: %s", strings.Join(tokens, " "))
		}

		envKey := strings.TrimSpace(tokens[0])
		envVal := strings.TrimSpace(tokens[1])

		envMap[envKey] = envVal
	}

	if err := envFileScanner.Err(); err != nil {
		return nil, err
	}

	return envMap, nil
}

func getOr(m map[string]string, key, def string) string {
	if v, ok := m[key]; ok && v != "" {
		return v
	}

	return def
}

func getDuration(m map[string]string, key string, def time.Duration) time.Duration {
	if v, ok := m[key]; ok && v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}
