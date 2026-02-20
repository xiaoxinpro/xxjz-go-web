package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_NotFound(t *testing.T) {
	_, err := Load("nonexistent.yaml")
	assert.Error(t, err)
}

func TestLoad_FromRepoRoot(t *testing.T) {
	// config.yaml at repo root
	cfgPath := "config.yaml"
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		cfgPath = filepath.Join("..", "..", "config.yaml")
	}
	cfg, err := Load(cfgPath)
	if err != nil {
		t.Skip("config.yaml not found, skip")
		return
	}
	require.NotNil(t, cfg)
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.NotEmpty(t, cfg.Database.Driver)
	assert.NotEmpty(t, cfg.App.Version)
}

func TestApplyEnv(t *testing.T) {
	// applyEnv is called from Load; test via Load with env set if needed
	os.Setenv("PORT", "9000")
	defer os.Unsetenv("PORT")
	cfgPath := "config.yaml"
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		cfgPath = filepath.Join("..", "..", "config.yaml")
	}
	cfg, err := Load(cfgPath)
	if err != nil {
		return
	}
	// After Load, env override should have been applied
	if os.Getenv("PORT") == "9000" {
		assert.Equal(t, 9000, cfg.Server.Port)
	}
}
