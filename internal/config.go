package internal

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

var (
	ErrConfigNotFound    = errors.New("config file not found")
	ErrInvalidProvider   = errors.New("invalid AI provider")
	ErrMissingAPIKey     = errors.New("API key is required")
	ErrConfigDirCreation = errors.New("failed to create config directory")
	ErrConfigFileWrite   = errors.New("failed to write config file")
	ErrConfigFileRead    = errors.New("failed to read config file")
	ErrInvalidConfigData = errors.New("invalid config data")
)

// AIProvider represents supported AI providers
type AIProvider string

const (
	ProviderClaude AIProvider = "claude"
	ProviderOpenAI AIProvider = "openai"
	ProviderGemini AIProvider = "gemini"
)

// AIConfig stores AI-related configuration
type AIConfig struct {
	Provider AIProvider `json:"provider"`
	APIKey   string     `json:"api_key"`
}

// GetConfigPath returns the path to the config file
func GetConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".git-tag-similarity", "config.json"), nil
}

// LoadConfig loads the AI configuration from disk
func LoadConfig() (*AIConfig, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrConfigNotFound
		}
		return nil, errors.Join(ErrConfigFileRead, err)
	}

	var config AIConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, errors.Join(ErrInvalidConfigData, err)
	}

	return &config, nil
}

// SaveConfig saves the AI configuration to disk
func SaveConfig(config *AIConfig) error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	// Create config directory if it doesn't exist
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return errors.Join(ErrConfigDirCreation, err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return errors.Join(ErrConfigFileWrite, err)
	}

	return nil
}

// Validate checks if the config is valid
func (c *AIConfig) Validate() error {
	switch c.Provider {
	case ProviderClaude, ProviderOpenAI, ProviderGemini:
		// Valid provider
	default:
		return errors.Join(ErrInvalidProvider, fmt.Errorf("unsupported provider: %s (supported: claude, openai, gemini)", c.Provider))
	}

	if c.APIKey == "" {
		return ErrMissingAPIKey
	}

	return nil
}

// ConfigCommandConfig holds the config command configuration
type ConfigCommandConfig struct {
	Provider string
	APIKey   string
}

// NewConfigCommandConfig parses the config command flags
func NewConfigCommandConfig(args []string) (ConfigCommandConfig, error) {
	config := ConfigCommandConfig{}

	configCmd := flag.NewFlagSet("config", flag.ExitOnError)
	configCmd.StringVar(&config.Provider, "provider", "claude", "AI provider (claude, openai, or gemini)")
	configCmd.StringVar(&config.APIKey, "api-key", "", "API key for the AI provider")

	configCmd.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: git-tag-similarity config [options]\n\n")
		fmt.Fprintf(os.Stderr, "Configure AI settings for report generation.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		configCmd.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  git-tag-similarity config -provider claude -api-key sk-ant-...\n")
		fmt.Fprintf(os.Stderr, "  git-tag-similarity config -provider openai -api-key sk-...\n")
		fmt.Fprintf(os.Stderr, "  git-tag-similarity config -provider gemini -api-key AIza...\n")
		fmt.Fprintf(os.Stderr, "\nSupported providers:\n")
		fmt.Fprintf(os.Stderr, "  claude    Anthropic Claude (default)\n")
		fmt.Fprintf(os.Stderr, "  openai    OpenAI GPT\n")
		fmt.Fprintf(os.Stderr, "  gemini    Google Gemini\n")
		fmt.Fprintf(os.Stderr, "\nNote: Your API key is stored in ~/.git-tag-similarity/config.json\n")
	}

	if err := configCmd.Parse(args); err != nil {
		return config, err
	}

	return config, nil
}

// Validate checks if the config command configuration is valid
func (c *ConfigCommandConfig) Validate() error {
	if c.Provider != "claude" && c.Provider != "openai" && c.Provider != "gemini" {
		return errors.Join(ErrInvalidProvider, fmt.Errorf("unsupported provider: %s (supported: claude, openai, gemini)", c.Provider))
	}

	if c.APIKey == "" {
		return ErrMissingAPIKey
	}

	return nil
}

// RunConfigCommand executes the config command
func RunConfigCommand(cmdConfig ConfigCommandConfig) error {
	if err := cmdConfig.Validate(); err != nil {
		return err
	}

	aiConfig := &AIConfig{
		Provider: AIProvider(cmdConfig.Provider),
		APIKey:   cmdConfig.APIKey,
	}

	if err := SaveConfig(aiConfig); err != nil {
		return err
	}

	fmt.Printf("Configuration saved successfully!\n")
	fmt.Printf("Provider: %s\n", aiConfig.Provider)
	fmt.Printf("API Key: %s...%s\n", aiConfig.APIKey[:8], aiConfig.APIKey[len(aiConfig.APIKey)-4:])

	return nil
}
