// run command - start Mock service
package cmd

import (
	"context"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"mock/internal/conf"
	"mock/internal/logger"
	"mock/internal/server"
	"os"
	"os/signal"
	"syscall"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Start Mock service",
	Long:  `Start Mock service using the specified mock.json configuration file.`,
	RunE:  runRun,
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().StringP("file", "f", "", "Configuration file")
	_ = runCmd.MarkFlagRequired("file")
}

// runRun executes the run command
func runRun(cmd *cobra.Command, args []string) error {
	filePath, _ := cmd.Flags().GetString("file")
	configs, err := conf.Load(filePath)
	if err != nil {
		return errors.Wrap(err, "Failed to load config")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg errgroup.Group
	for _, cfg := range configs {
		wg.Go(func() error {
			return server.StartServer(ctx, cfg)
		})
	}

	return wait(cancel, &wg)
}

func wait(cancel context.CancelFunc, wg *errgroup.Group) error {
	sign := make(chan os.Signal, 1)
	signal.Notify(sign, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sign
		cancel()
	}()

	err := wg.Wait()

	// Close logger to ensure all pending logs are written
	_ = logger.Close()
	return err
}
