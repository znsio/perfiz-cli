package cmd

import (
	"github.com/spf13/cobra"
	"github.com/znsio/perfiz-cli/common/constants"
	"github.com/znsio/perfiz-cli/common/path"
	"log"
	"os"
	"os/exec"
)

func init() {
	rootCmd.AddCommand(cmdReset)
}

var cmdReset = &cobra.Command{
	Use:   "reset",
	Short: "removes project specific grafana and prometheus data",
	Long:  `removes <your project folder>/perfiz/*_data to reset Grafana, InfluxDB and Prometheus specific to that project`,
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		dockerNetworkCheck := exec.Command("docker", "network", "inspect", "perfiz-network")

		_, dockerNetworkCheckError := dockerNetworkCheck.Output()
		if dockerNetworkCheckError == nil {
			log.Fatalln("Perfiz Containers seem to be running. Please run 'stop' command before running 'reset'.")
		}

		if !path.IsDir(constants.PERFIZ_FOLDER) {
			log.Fatalln("Could not find perfiz folder, please run 'reset' command inside your project where the perfiz folder exists.")
		}

		for _, dataDirPath := range []string{constants.PERFIZ_FOLDER + "/grafana_data", constants.PERFIZ_FOLDER + "/influxdb_data", constants.PERFIZ_FOLDER + "/prometheus_data", constants.PERFIZ_FOLDER + "/gatling_data"} {
			log.Println("Deleting " + dataDirPath)
			os.RemoveAll(dataDirPath)
		}
	},
}