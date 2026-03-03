package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

type sampleConfig struct {
	NATSURL    string `mapstructure:"nats_url" env:"NATS_URL"`
	PluginsDir string `mapstructure:"plugins_dir" env:"PLUGINS_DIR" default:"plugins"`
	Log        struct {
		Format string `mapstructure:"format" env:"LOG_FORMAT" default:"json"`
	} `mapstructure:"log"`
}

func TestLoad_AppliesDefaults(t *testing.T) {
	t.Setenv("NATS_URL", "nats://localhost:4222")
	t.Setenv("PLUGINS_DIR", "")
	t.Setenv("LOG_FORMAT", "")

	tmp := t.TempDir()
	oldWD, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmp))
	t.Cleanup(func() { _ = os.Chdir(oldWD) })

	var cfg sampleConfig
	require.NoError(t, Load(&cfg))
	require.Equal(t, "nats://localhost:4222", cfg.NATSURL)
	require.Equal(t, "plugins", cfg.PluginsDir)
	require.Equal(t, "json", cfg.Log.Format)
}

func TestLoad_ReadsDotEnvFile(t *testing.T) {
	tmp := t.TempDir()
	oldWD, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmp))
	t.Cleanup(func() { _ = os.Chdir(oldWD) })

	dotEnv := "NATS_URL=nats://from-dotenv:4222\nPLUGINS_DIR=plugins-live\nLOG_FORMAT=pretty\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmp, ".env"), []byte(dotEnv), 0o644))

	var cfg sampleConfig
	require.NoError(t, Load(&cfg))
	require.Equal(t, "nats://from-dotenv:4222", cfg.NATSURL)
	require.Equal(t, "plugins-live", cfg.PluginsDir)
	require.Equal(t, "pretty", cfg.Log.Format)
}

func TestLoad_MissingRequiredReturnsError(t *testing.T) {
	tmp := t.TempDir()
	oldWD, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmp))
	t.Cleanup(func() { _ = os.Chdir(oldWD) })

	var cfg sampleConfig
	err = Load(&cfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "NATS_URL")
}
