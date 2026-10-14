package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestNormalizeThemeMigratesLegacyDefault(t *testing.T) {
	for _, theme := range []string{"", "default", " DEFAULT "} {
		if got := NormalizeTheme(theme); got != DefaultTheme {
			t.Errorf("NormalizeTheme(%q) = %q, want %q", theme, got, DefaultTheme)
		}
	}
}

func TestNormalizeThemePreservesCustomTheme(t *testing.T) {
	const custom = "/tmp/my-gallery-theme"
	if got := NormalizeTheme("  " + custom + "  "); got != custom {
		t.Fatalf("NormalizeTheme() = %q, want %q", got, custom)
	}
}

func TestSecureConfigPermissions(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte("gallery: {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	viper.SetConfigFile(path)

	if err := secureConfigPermissions(); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("config permissions = %o, want 600", got)
	}
}

func TestLoadConfigRejectsMalformedFile(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)
	path := filepath.Join(t.TempDir(), "broken.yaml")
	if err := os.WriteFile(path, []byte("gallery: [\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	viper.SetConfigFile(path)

	if _, err := LoadConfig(); err == nil || !strings.Contains(err.Error(), "read configuration") {
		t.Fatalf("expected malformed config error, got %v", err)
	}
}

func TestAIRequiresExplicitOptInAndPersists(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)
	previous := Config
	Config = &Configuration{UI: UIConfiguration{AIEnabled: true}}
	t.Cleanup(func() { Config = previous })
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte("ui:\n  geminiapikey: saved-test-key\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	viper.SetConfigFile(path)
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.UI.AIEnabled || cfg.UI.GeminiApiKey != "saved-test-key" {
		t.Fatal("an existing API key must not implicitly enable AI")
	}
	for _, enabled := range []bool{true, false} {
		cfg.UI.AIEnabled = enabled
		if err := cfg.Save(); err != nil {
			t.Fatal(err)
		}
		viper.Reset()
		viper.SetConfigFile(path)
		Config = &Configuration{}
		cfg, err = LoadConfig()
		if err != nil {
			t.Fatal(err)
		}
		if cfg.UI.AIEnabled != enabled || cfg.UI.GeminiApiKey != "saved-test-key" {
			t.Fatalf("AI preference did not survive reload: enabled = %v", cfg.UI.AIEnabled)
		}
	}
}

func TestNewConfigurationDefaultsAIOff(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)
	previous := Config
	Config = &Configuration{}
	t.Cleanup(func() { Config = previous })
	viper.SetConfigFile(filepath.Join(t.TempDir(), "new.yaml"))
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.UI.AIEnabled {
		t.Fatal("AI must be disabled in a new configuration")
	}
}
