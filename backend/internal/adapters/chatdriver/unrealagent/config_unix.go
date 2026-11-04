//go:build darwin || linux

package unrealagent

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

func readProviderConfig(path string) (providerConfig, error) {
	file, err := os.Open(path) //nolint:gosec // path is supplied only by AO's hidden child command
	if err != nil {
		return providerConfig{}, fmt.Errorf("open Unreal Agent provider config: %w", err)
	}
	defer func() { _ = file.Close() }()
	var cfg providerConfig
	decoder := json.NewDecoder(io.LimitReader(file, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&cfg); err != nil {
		return providerConfig{}, fmt.Errorf("decode Unreal Agent provider config: %w", err)
	}
	if err := validateProviderConfig(cfg); err != nil {
		return providerConfig{}, err
	}
	return cfg, nil
}
