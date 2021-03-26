package main

import (
	"github.com/otiai10/copy"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	PERFIZ_HOME_ENV_VARIABLE     = "PERFIZ_HOME"
	PERFIZ_YML                   = "perfiz.yml"
	GRAFANA_DASHBOARDS_DIRECTORY = "./perfiz/dashboards"
	PROMETHEUS_CONFIG            = "./perfiz/prometheus/prometheus.yml"
)

type PerfizConfig struct {
	KarateFeaturesDir string `yaml:"karateFeaturesDir"`
}

func main() {
	var cmdStart = &cobra.Command{
		Use:   "start",
		Short: "Start Perfiz",
		Long:  `Start Perfiz Docker Containers and run load test`,
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			workingDir, _ := os.Getwd()
			log.Println("Starting Perfiz...")
			perfizHome := os.Getenv(PERFIZ_HOME_ENV_VARIABLE)
			if len(perfizHome) == 0 {
				log.Fatalln("Please set " + PERFIZ_HOME_ENV_VARIABLE + " environment variable")
			} else {
				log.Println(PERFIZ_HOME_ENV_VARIABLE + ": " + perfizHome)
			}
			_, perfizYmlErr := os.Open(PERFIZ_YML)
			if perfizYmlErr != nil {
				log.Fatalln(PERFIZ_YML+" not found. Please see https://github.com/znsio/perfiz for instructions", perfizYmlErr)
			}
			perfizConfig := &PerfizConfig{}
			b, _ := ioutil.ReadFile(PERFIZ_YML)
			configParseError := yaml.Unmarshal(b, perfizConfig)
			if configParseError != nil {
				log.Fatal(configParseError)
			}
			karateFeaturesDir := workingDir + "/" + perfizConfig.KarateFeaturesDir
			if !IsDir(karateFeaturesDir) {
				log.Fatalln("Configuration error in perfiz.yml. karateFeaturesDir: " + perfizConfig.KarateFeaturesDir + ". " + karateFeaturesDir + " is not a directory. Please note that karateFeaturesDir has to be relative to perfiz.yml location.")
			}
			if IsDir(GRAFANA_DASHBOARDS_DIRECTORY) {
				log.Println("Copying Grafana Dashboard jsons in " + GRAFANA_DASHBOARDS_DIRECTORY)
				copy.Copy(GRAFANA_DASHBOARDS_DIRECTORY, perfizHome+"/prometheus-metrics-monitor/grafana/dashboards")
			}
			_, prometheusConfigErr := os.Open(PROMETHEUS_CONFIG)
			if prometheusConfigErr == nil {
				log.Println("Copying prometheus.yml in " + PROMETHEUS_CONFIG)
				copy.Copy(PROMETHEUS_CONFIG, perfizHome+"/prometheus-metrics-monitor/prometheus/prometheus.yml")
			}
			log.Println("Starting Perfiz Docker Containers...")
			dockerComposeUp := exec.Command("docker-compose", "-f", perfizHome+"/docker-compose.yml", "up", "-d")
			dockerComposeUpOutput, dockerComposeUpError := dockerComposeUp.Output()
			if dockerComposeUpError != nil {
				log.Fatalln(dockerComposeUpError.Error())
			}
			log.Println(string(dockerComposeUpOutput))
			log.Println("Navigate to http://localhost:3000 for Grafana")

			perfizMavenRepo := perfizHome + "/.m2"
			if IsDir(perfizMavenRepo) {
				log.Println(perfizMavenRepo + " available. Skipping Maven Dependency Download.")
			} else {
				log.Println(perfizMavenRepo + " does not exist. Maven dependencies will be run downloaded. This may take a while...")
			}

			dockerRun := exec.Command("docker", "run", "--rm", "--name", "perfiz-gatling",
				"-v", perfizMavenRepo+":/root/.m2",
				"-v", perfizHome+":/usr/src/performance-testing",
				"-v", karateFeaturesDir+":/usr/src/karate-features",
				"-v", workingDir+"/"+PERFIZ_YML+":/usr/src/perfiz.yml",
				"-e", "KARATE_FEATURES=/usr/src/karate-features",
				"-w", "/usr/src/performance-testing",
				"--network", "perfiz-network",
				"maven:3.6-jdk-8", "mvn", "clean", "test-compile", "gatling:test", "-DPERFIZ=/usr/src/perfiz.yml",
			)
			log.Println("Starting Gatling Tests...")
			log.Println(dockerRun)
			dockerRunOutput, dockerRunError := dockerRun.Output()
			if dockerRunError != nil {
				log.Fatalln(dockerRunError.Error())
			}
			log.Println(string(dockerRunOutput))
		},
	}

	var cmdStop = &cobra.Command{
		Use:   "stop",
		Short: "Stop Perfiz",
		Long:  `Stop Perfiz Docker Containers`,
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			log.Println("Stopping Perfiz...")
			perfizHome := os.Getenv(PERFIZ_HOME_ENV_VARIABLE)
			if len(perfizHome) == 0 {
				log.Fatalln("Please set " + PERFIZ_HOME_ENV_VARIABLE + " environment variable")
			} else {
				log.Println(PERFIZ_HOME_ENV_VARIABLE + ": " + perfizHome)
			}
			dockerComposeDown := exec.Command("docker-compose", "-f", perfizHome+"/docker-compose.yml", "down")
			dockerComposeDownOutput, dockerComposeDownError := dockerComposeDown.Output()
			if dockerComposeDownError != nil {
				log.Fatalln(dockerComposeDownError.Error())
			}
			log.Println(string(dockerComposeDownOutput))
		},
	}

	var rootCmd = &cobra.Command{Use: "perfiz-cli"}
	rootCmd.AddCommand(cmdStart, cmdStop)
	rootCmd.Execute()
}

func IsDir(pathFile string) bool {
	if pathAbs, err := filepath.Abs(pathFile); err != nil {
		return false
	} else if fileInfo, err := os.Stat(pathAbs); os.IsNotExist(err) || !fileInfo.IsDir() {
		return false
	}

	return true
}
