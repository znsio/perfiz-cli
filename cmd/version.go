package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/znsio/perfiz-cli/common/constants"
	"github.com/znsio/perfiz-cli/common/version"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Perfiz",
	Long:  `All software has versions. This is Perfiz's`,
	Run: func(cmd *cobra.Command, args []string) {
		var perfizVersion = version.GetPerfizVersion()
		fmt.Println("********** PERFIZ VERSION **********")
		fmt.Println("perfiz " + perfizVersion)
		fmt.Println("perfiz-cli " + constants.PERFIZ_CLI_VERSION)
		fmt.Println("************************************")
	},
}