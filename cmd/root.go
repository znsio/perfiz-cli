package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "perfiz",
	Short: "A Dockerised Performance Test Setup",
	Long: `A Dockerised API Performance Test Setup based on Gatling with Grafana Dashboards and Prometheus Monitoring.
                Complete documentation is available at https://perfiz.com`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
