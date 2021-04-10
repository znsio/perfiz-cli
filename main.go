package main

import (
	"bufio"
	"fmt"
	"github.com/otiai10/copy"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	PERFIZ_HOME_ENV_VARIABLE     = "PERFIZ_HOME"
	PERFIZ_YML                   = "perfiz.yml"
	GRAFANA_DASHBOARDS_DIRECTORY = "./perfiz/dashboards"
	PROMETHEUS_CONFIG            = "./perfiz/prometheus/prometheus.yml"
)

type PerfizConfig struct {
	KarateFeaturesDir     string `yaml:"karateFeaturesDir"`
	GatlingSimulationsDir string `yaml:"gatlingSimulationsDir"`
}

func main() {
	var cmdStart = &cobra.Command{
		Use:   "start",
		Short: "Start Perfiz Monitoring Stack",
		Long:  `Start Grafana, Prometheus and other Monitoring Stack Docker Containers`,
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			log.Println("Starting Perfiz...")
			perfizHome := getEnvVariable(PERFIZ_HOME_ENV_VARIABLE)
			checkIfCommandExists("docker")
			checkIfCommandExists("docker-compose")

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
			dockerComposeUpOutput, dockerComposeUpError := dockerComposeUp.CombinedOutput()
			if dockerComposeUpError != nil {
				log.Println(fmt.Sprint(dockerComposeUpError) + ": " + string(dockerComposeUpOutput))
				log.Fatalln(dockerComposeUpError.Error())
			}
			log.Println(string(dockerComposeUpOutput))
			log.Println("Navigate to http://localhost:3000 for Grafana")
		},
	}

	var cmdTest = &cobra.Command{
		Use:   "test",
		Short: "Run Gatling Performance Test",
		Long:  `Run Gatling Performance Tests as per the configuration in perfiz.yml`,
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			workingDir, _ := os.Getwd()
			perfizHome := getEnvVariable(PERFIZ_HOME_ENV_VARIABLE)
			checkIfCommandExists("docker")
			checkIfCommandExists("docker-compose")

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
			gatlingSimulationsDir := getGatlingSimulationsDir(workingDir, perfizConfig)
			if gatlingSimulationsDir != "" && !IsDir(gatlingSimulationsDir) {
				log.Fatalln("Configuration error in perfiz.yml. gatlingSimulationsDir: " + perfizConfig.GatlingSimulationsDir + ". " + gatlingSimulationsDir + " is not a directory. Please note that gatlingSimulationsDir has to be relative to perfiz.yml location.")
			}

			libRegEx, e := regexp.Compile("^*.scala")
			if e != nil {
				log.Fatal(e)
			}

			filepath.Walk(perfizHome+"/src/test/scala/", func(path string, info os.FileInfo, err error) error {
				if err == nil && libRegEx.MatchString(info.Name()) && !strings.Contains(info.Name(), "Perfiz") {
					log.Println("Removing " + info.Name())
					os.Remove(path)
				}
				return nil
			})

			if gatlingSimulationsDir != "" {
				log.Println("Copying Gatling Simulations in " + gatlingSimulationsDir)
				onlyScalaSimulationFiles := copy.Options{
					Skip: func(src string) (bool, error) {
						return !IsDir(src) && !strings.HasSuffix(src, ".scala"), nil
					},
				}
				copy.Copy(gatlingSimulationsDir, perfizHome+"/src/test/scala", onlyScalaSimulationFiles)
			}

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
			dockerRunOutput, _ := dockerRun.StdoutPipe()
			dockerRunError, _ := dockerRun.StderrPipe()

			dockerRun.Start()

			logStreamingOutput(dockerRunOutput)

			logStreamingOutput(dockerRunError)

			dockerRun.Wait()
		},
	}

	var cmdStop = &cobra.Command{
		Use:   "stop",
		Short: "Stop Perfiz Monitoring Stack",
		Long:  `Stop all Perfiz related Docker Containers`,
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
	rootCmd.AddCommand(cmdStart, cmdTest, cmdStop)
	rootCmd.Execute()
}

func getGatlingSimulationsDir(workingDir string, config *PerfizConfig) string {
	if config.GatlingSimulationsDir == "" {
		return ""
	}
	return workingDir + "/" + config.GatlingSimulationsDir
}

func logStreamingOutput(output io.ReadCloser) {
	outputScanner := bufio.NewScanner(output)
	outputScanner.Split(bufio.ScanLines)
	for outputScanner.Scan() {
		log.Println(outputScanner.Text())
	}
}

func checkIfCommandExists(command string) {
	path, err := exec.LookPath(command)
	if err != nil {
		log.Fatalln(command+" not found, please install. Error: ", err)
	}

	log.Println(command + " command located: " + path)
}

func getEnvVariable(envVariableName string) string {
	envVariable := os.Getenv(envVariableName)
	if len(envVariable) == 0 {
		log.Fatalln("Please set " + envVariableName + " environment variable")
	}
	log.Println(envVariableName + ": " + envVariable)
	return envVariable
}

func IsDir(pathFile string) bool {
	if pathAbs, err := filepath.Abs(pathFile); err != nil {
		return false
	} else if fileInfo, err := os.Stat(pathAbs); os.IsNotExist(err) || !fileInfo.IsDir() {
		return false
	}

	return true
}
