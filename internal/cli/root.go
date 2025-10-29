package cli

import (
	"github.com/spf13/cobra"
	"github.com/sage-x-project/blockchain-indexer/internal/cli/cmd"
)

var (
	version string
	commit  string
	date    string
)

// SetVersion sets the version information
func SetVersion(v, c, d string) {
	version = v
	commit = c
	date = d
}

// Execute executes the root command
func Execute() error {
	rootCmd := &cobra.Command{
		Use:   "blockchain-indexer",
		Short: "A high-performance blockchain indexer",
		Long: `Blockchain Indexer is a production-ready indexer for multiple blockchains.
It supports EVM-compatible chains, Solana, Cosmos, and more.

Features:
  - Multi-chain support
  - Real-time indexing
  - GraphQL, gRPC, and REST APIs
  - Event-driven architecture
  - Gap detection and recovery`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// Add commands
	rootCmd.AddCommand(cmd.NewServerCmd())
	rootCmd.AddCommand(cmd.NewIndexCmd())
	rootCmd.AddCommand(cmd.NewVersionCmd(version, commit, date))
	rootCmd.AddCommand(cmd.NewConfigCmd())

	return rootCmd.Execute()
}
