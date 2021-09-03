package cmd

import (
	"github.com/spf13/cobra"
	"github.com/znsio/perfiz-cli/common/constants"
	"log"
	"os"
	"os/exec"
)

func init() {
	rootCmd.AddCommand(cmdStop)
}

var cmdStop = &cobra.Command{
	Use:   "stop",
	Short: "Stop Perfiz Monitoring Stack",
	Long:  `Stop all Perfiz related Docker Containers`,
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("Stopping Perfiz...")
		perfizHome := os.Getenv(constants.PERFIZ_HOME_ENV_VARIABLE)
		if len(perfizHome) == 0 {
			log.Fatalln("Please set " + constants.PERFIZ_HOME_ENV_VARIABLE + " environment variable")
		} else {
			log.Println(constants.PERFIZ_HOME_ENV_VARIABLE + ": " + perfizHome)
		}
		dockerComposeDown := exec.Command("docker-compose", "--file", perfizHome+"/docker-compose.yml", "--env-file", perfizHome+constants.DOCKER_COMPOSE_ENV_FILE, "down")
		log.Println("Docker Compose Command: ")
		log.Println(dockerComposeDown)
		dockerComposeDownOutput, dockerComposeDownError := dockerComposeDown.Output()
		if dockerComposeDownError != nil {
			log.Fatalln(dockerComposeDownError.Error())
		}
		log.Println(string(dockerComposeDownOutput))
	},
}
