package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewVersionCmd creates a version command
func NewVersionCmd(version, commit, date string) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Long:  "Print the version, commit hash, and build date of the blockchain indexer",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Blockchain Indexer\n")
			fmt.Printf("  Version:    %s\n", version)
			fmt.Printf("  Commit:     %s\n", commit)
			fmt.Printf("  Built:      %s\n", date)
		},
	}
}
