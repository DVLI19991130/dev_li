// commands 包 - 定义所有 CLI 命令
package cmd

import (
	"github.com/spf13/cobra"
	"os"
)

// Root command
var rootCmd = &cobra.Command{
	Use:     "mock",
	Short:   "mock is a high-performance interface Mock CLI tool",
	Version: "1.0.0",
	Long: `mock is a high-performance interface Mock CLI tool implemented in Go,
used for mocking or proxying target interfaces during project interface performance testing,
with dynamic configuration support via mock.json.`,
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
