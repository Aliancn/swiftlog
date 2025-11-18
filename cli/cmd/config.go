package cmd

import (
	"fmt"

	"github.com/aliancn/swiftlog/cli/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage SwiftLog CLI configuration",
	Long:  `Configure authentication tokens and server settings for the SwiftLog CLI.`,
}

var configSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set configuration values",
	Long:  `Set configuration values such as API token and server address.`,
	RunE:  runConfigSet,
}

var configGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get configuration values",
	Long:  `Display current configuration values.`,
	RunE:  runConfigGet,
}

var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Show configuration file path",
	Long:  `Display the path to the configuration file.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(config.GetConfigPath())
	},
}

var (
	setToken  string
	setServer string
)

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configPathCmd)

	// Flags for config set
	configSetCmd.Flags().StringVar(&setToken, "token", "", "API token")
	configSetCmd.Flags().StringVar(&setServer, "server", "", "Server address (e.g., localhost:50051)")
}

func runConfigSet(cmd *cobra.Command, args []string) error {
	// Load existing config or create new one
	cfg, err := config.Load()
	if err != nil {
		cfg = &config.Config{
			ServerAddr: "localhost:50051",
		}
	}

	// Update values if provided
	if setToken != "" {
		cfg.Token = setToken
		fmt.Println("✓ Token updated")
	}
	if setServer != "" {
		cfg.ServerAddr = setServer
		fmt.Println("✓ Server address updated")
	}

	// Save config
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("\nConfiguration saved to: %s\n", config.GetConfigPath())
	return nil
}

func runConfigGet(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println("Current configuration:")
	fmt.Printf("  Server Address: %s\n", cfg.ServerAddr)
	if cfg.Token != "" {
		// Mask token for security
		maskedToken := cfg.Token
		if len(maskedToken) > 8 {
			maskedToken = maskedToken[:4] + "..." + maskedToken[len(maskedToken)-4:]
		}
		fmt.Printf("  API Token:      %s\n", maskedToken)
	} else {
		fmt.Println("  API Token:      (not set)")
	}
	fmt.Printf("\nConfig file: %s\n", config.GetConfigPath())

	return nil
}
