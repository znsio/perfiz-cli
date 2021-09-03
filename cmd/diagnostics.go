package cmd

import (
	"github.com/spf13/cobra"
	"github.com/znsio/perfiz-cli/common/constants"
	"github.com/znsio/perfiz-cli/common/environment"
	"github.com/znsio/perfiz-cli/common/version"
	"log"
	"os"
	"runtime"
)

func init() {
	rootCmd.AddCommand(cmdDiagnostics)
}

var cmdDiagnostics = &cobra.Command{
	Use:   "diagnostics",
	Short: "gathers setup information to report issues",
	Long:  `Gathers setup information to report issues.`,
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("*************** RUNNING DIAGNOSTICS ******************")
		perfizVersion := version.GetPerfizVersion()
		log.Println("Perfiz Version: " + perfizVersion)
		log.Println("Perfiz Cli Version: " + constants.PERFIZ_CLI_VERSION)
		log.Println("Docker version: " + environment.GetCommandVersion("docker"))
		log.Println("docker-compose version: " + environment.GetCommandVersion("docker-compose"))
		log.Println("OS: " + runtime.GOOS)
		log.Println("Arch: " + runtime.GOARCH)
		perfizFolderStats, perfizFolderStatsErr := os.Stat(constants.PERFIZ_FOLDER)
		if perfizFolderStatsErr == nil {
			log.Println("Perfiz folder permissions: " + perfizFolderStats.Mode().String())
		} else {
			log.Println("Error reading permissions of Perfiz Folder. " + perfizFolderStatsErr.Error())
		}
		log.Println("************* DIAGNOSTICS COMPLETED ******************")
	},
}