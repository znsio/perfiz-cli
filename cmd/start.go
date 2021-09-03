package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/znsio/perfiz-cli/common/constants"
	env "github.com/znsio/perfiz-cli/common/environment"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
)

func init() {
	rootCmd.AddCommand(cmdStart)
}

var cmdStart = &cobra.Command{
	Use:   "start",
	Short: "Start Perfiz Monitoring Stack",
	Long:  `Start Grafana, Prometheus and other Monitoring Stack Docker Containers`,
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("Starting Perfiz...")
		perfizHome := env.GetEnvVariable(constants.PERFIZ_HOME_ENV_VARIABLE)
		env.CheckIfCommandExists("docker", constants.DOCKER_MAJOR_VERSION, constants.DOCKER_MINOR_VERSION)
		env.CheckIfCommandExists("docker-compose", constants.DOCKER_COMPOSE_MAJOR_VERSION, constants.DOCKER_COMPOSE_MINOR_VERSION)

		workingDir, _ := os.Getwd()
		log.Println("Writing PROJECT_DIR=" + workingDir + " to docker-compose env file: " + perfizHome + constants.DOCKER_COMPOSE_ENV_FILE)
		err := ioutil.WriteFile(perfizHome+constants.DOCKER_COMPOSE_ENV_FILE, []byte("PROJECT_DIR="+workingDir), 0755)

		if err != nil {
			log.Println("Error writing docker-compose .env: " + perfizHome + constants.DOCKER_COMPOSE_ENV_FILE)
			log.Fatalln(err)
		}

		log.Println("Starting Perfiz Docker Containers...")
		dockerComposeUp := exec.Command("docker-compose", "--file", perfizHome+"/docker-compose.yml", "--env-file", perfizHome+constants.DOCKER_COMPOSE_ENV_FILE, "up", "-d")
		log.Println("Docker Compose Command: ")
		log.Println(dockerComposeUp)
		dockerComposeUpOutput, dockerComposeUpError := dockerComposeUp.CombinedOutput()
		if dockerComposeUpError != nil {
			log.Println(fmt.Sprint(dockerComposeUpError) + ": " + string(dockerComposeUpOutput))
			log.Fatalln(dockerComposeUpError.Error())
		}
		log.Println(string(dockerComposeUpOutput))
		log.Println("Navigate to http://localhost:3000 for Grafana")
	},
}