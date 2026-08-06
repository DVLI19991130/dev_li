// check command - validate mock.json configuration file
package cmd

import (
	"mock/internal/conf"

	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Validate mock.json configuration file",
	Long:  `Validate the format and content of the specified mock.json configuration file.`,
	RunE:  runCheck,
}

func init() {
	rootCmd.AddCommand(checkCmd)

	checkCmd.Flags().StringP("file", "f", "mock.json", "Configuration file")
}

// runCheck executes the check command
func runCheck(cmd *cobra.Command, args []string) error {
	filePath, _ := cmd.Flags().GetString("file")
	_, err := conf.Load(filePath)
	return err
}
