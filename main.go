package main

import (
	"bufio"
	"errors"
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
	"strconv"
	"strings"
)

const (
	PERFIZ_HOME_ENV_VARIABLE     = "PERFIZ_HOME"
	DEFAULT_CONFIG_FILE          = "perfiz.yml"
	GATLING_CONF                 = "gatling.conf"
	GATLING_CONF_PATH            = "./perfiz/gatling/"
	GRAFANA_DASHBOARDS_DIRECTORY = "./perfiz/dashboards"
	PROMETHEUS_CONFIG_DIR        = "./perfiz/prometheus"
	PROMETHEUS_CONFIG            = PROMETHEUS_CONFIG_DIR + "/prometheus.yml"
	DOCKER_MAJOR_VERSION         = 20
	DOCKER_MINOR_VERSION         = 10
)

type PerfizConfig struct {
	KarateFeaturesDir     string `yaml:"karateFeaturesDir"`
	KarateEnv             string `yaml:"karateEnv"`
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
			checkIfCommandExists("docker", DOCKER_MAJOR_VERSION, DOCKER_MINOR_VERSION)

			workingDir, _ := os.Getwd()
			log.Println("Writing working directory to docker-compose .env: " + workingDir)
			ioutil.WriteFile(perfizHome+"/.env", []byte("PROJECT_DIR="+workingDir), 0755)

			log.Println("Starting Perfiz Docker Containers...")
			dockerComposeUp := exec.Command("docker", "compose", "-f", perfizHome+"/docker-compose.yml", "up", "-d")
			dockerComposeUpOutput, dockerComposeUpError := dockerComposeUp.CombinedOutput()
			if dockerComposeUpError != nil {
				log.Println(fmt.Sprint(dockerComposeUpError) + ": " + string(dockerComposeUpOutput))
				log.Fatalln(dockerComposeUpError.Error())
			}
			log.Println(string(dockerComposeUpOutput))
			log.Println("Navigate to http://localhost:3000 for Grafana")
		},
	}

	var cmdInit = &cobra.Command{
		Use:   "init",
		Short: "Add Perfiz Config Templates and Dirs",
		Long: `Add Perfiz Config YML template, Directories for Grafana Dashboards,
                Prometheus Configs and update .gitignore`,
		Args: cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			log.Println("Staring Init")
			perfizHome := getEnvVariable(PERFIZ_HOME_ENV_VARIABLE)
			addTemplateIfMissing(perfizHome, DEFAULT_CONFIG_FILE, "./")
			addTemplateIfMissing(perfizHome, GATLING_CONF, GATLING_CONF_PATH)
			if !IsDir(GRAFANA_DASHBOARDS_DIRECTORY) {
				log.Println("Creating Grafana Dashboard dir " + GRAFANA_DASHBOARDS_DIRECTORY + ". Add Grafana Dashboard JSONs here.")
				os.MkdirAll(GRAFANA_DASHBOARDS_DIRECTORY, 0755)
				log.Println("Adding sample dashboard json " + GRAFANA_DASHBOARDS_DIRECTORY + "/dashboard.json as a reference to get you started.")
				copy.Copy(perfizHome+"/templates/dashboard.json", GRAFANA_DASHBOARDS_DIRECTORY+"/dashboard.json")
			} else {
				log.Println(GRAFANA_DASHBOARDS_DIRECTORY + " is already present. Skipping.")
			}
			_, prometheusConfigErr := os.Open(PROMETHEUS_CONFIG)
			if prometheusConfigErr != nil {
				log.Println("Creating prometheus.yml template in " + PROMETHEUS_CONFIG + ". Add scrape configs to this file.")
				os.MkdirAll(PROMETHEUS_CONFIG_DIR, 0755)
				copy.Copy(perfizHome+"/templates/prometheus.yml", PROMETHEUS_CONFIG)
			} else {
				log.Println(PROMETHEUS_CONFIG + " is already present. Skipping.")
			}
			log.Println("Init Completed")
			log.Println("Please add below line to your .gitignore to avoid checking in Prometheus and Grafana Data to version control")
			log.Println("perfiz/*_data")
		},
	}

	var cmdTest = &cobra.Command{
		Use:   "test [perfiz config file name]",
		Short: "Run Gatling Performance Test",
		Long:  `Run Gatling Performance Tests as per the configuration in perfiz.yml`,
		Args: func(cmd *cobra.Command, args []string) error {
			_, perfizYmlErr := os.Open(DEFAULT_CONFIG_FILE)
			if len(args) < 1 {
				if perfizYmlErr != nil {
					return errors.New("Default Config: " + DEFAULT_CONFIG_FILE + " not found. Please create " + DEFAULT_CONFIG_FILE + " or provide name of config file as argument. Please see https://github.com/znsio/perfiz for instructions and / or run 'init' command and perfiz will add a config file template to help you get started.")
				} else {
					return nil
				}
			}
			_, customPerfizYmlErr := os.Open(args[0])
			if customPerfizYmlErr != nil {
				return errors.New("Config: " + args[0] + " not found.")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			workingDir, _ := os.Getwd()
			perfizHome := getEnvVariable(PERFIZ_HOME_ENV_VARIABLE)
			checkIfCommandExists("docker", DOCKER_MAJOR_VERSION, DOCKER_MINOR_VERSION)

			var configFile string
			if len(args) == 1 {
				configFile = args[0]
			} else {
				configFile = DEFAULT_CONFIG_FILE
			}
			log.Println("Perfiz Config File: " + configFile)
			perfizConfig := &PerfizConfig{}
			b, _ := ioutil.ReadFile(configFile)
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

			_, gatlingConfErr := os.Open(GATLING_CONF_PATH + GATLING_CONF)
			if gatlingConfErr == nil {
				log.Println("Copying Gatling Configuration " + GATLING_CONF_PATH + GATLING_CONF)
				copy.Copy(GATLING_CONF_PATH+GATLING_CONF, perfizHome+"/src/test/resources/"+GATLING_CONF)
			}

			perfizMavenRepo := perfizHome + "/.m2"
			if IsDir(perfizMavenRepo) {
				log.Println(perfizMavenRepo + " available. Skipping Maven Dependency Download.")
			} else {
				log.Println(perfizMavenRepo + " does not exist. Maven dependencies will be run downloaded. This may take a while...")
			}

			dockerNetworkCheck := exec.Command("docker", "network", "inspect", "perfiz-network")
			log.Println("Running checks...")

			_, dockerNetworkCheckError := dockerNetworkCheck.Output()
			if dockerNetworkCheckError != nil {
				log.Fatalln("Error locating docker network perfiz-network. Try running perfiz 'start' command before running 'test'.")
			}

			log.Println("All checks done.")

			dockerCommandArguments := []string{"run", "--rm", "--name", "perfiz-gatling",
				"-v", perfizMavenRepo + ":/root/.m2",
				"-v", perfizHome + ":/usr/src/performance-testing",
				"-v", karateFeaturesDir + ":/usr/src/karate-features",
				"-v", workingDir + "/" + configFile + ":/usr/src/perfiz.yml",
				"-e", "KARATE_FEATURES=/usr/src/karate-features",
				"-w", "/usr/src/performance-testing",
				"--network", "perfiz-network",
				"maven:3.6-jdk-8", "mvn", "clean", "test-compile", "gatling:test", "-DPERFIZ=/usr/src/perfiz.yml"}

			if perfizConfig.KarateEnv != "" {
				log.Println("Setting karate.env to " + perfizConfig.KarateEnv)
				dockerCommandArguments = append(dockerCommandArguments, "-Dkarate.env="+perfizConfig.KarateEnv)
			}

			dockerRun := exec.Command("docker", dockerCommandArguments...)
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
			dockerComposeDown := exec.Command("docker", "compose", "-f", perfizHome+"/docker-compose.yml", "down")
			dockerComposeDownOutput, dockerComposeDownError := dockerComposeDown.Output()
			if dockerComposeDownError != nil {
				log.Fatalln(dockerComposeDownError.Error())
			}
			log.Println(string(dockerComposeDownOutput))
		},
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

			if !IsDir("./perfiz") {
				log.Fatalln("Could not find perfiz folder, please run 'reset' command inside your project where the perfiz folder exists.")
			}

			for _, dataDirPath := range []string{"./perfiz/grafana_data", "./perfiz/influxdb_data", "./perfiz/prometheus_data"} {
				log.Println("Deleting " + dataDirPath)
				os.RemoveAll(dataDirPath)
			}
		},
	}

	var rootCmd = &cobra.Command{Use: "perfiz-cli"}
	rootCmd.AddCommand(cmdInit, cmdStart, cmdTest, cmdStop, cmdReset)
	rootCmd.Execute()
}

func addTemplateIfMissing(perfizHome string, filename string, path string) {
	filePath := path + filename
	_, fileOpenErr := os.Open(filePath)
	if fileOpenErr != nil {
		log.Println(filePath + " not found. Adding template.")
		copy.Copy(perfizHome+"/templates/"+filename, filePath)
	} else {
		log.Println(filePath + " is already present. Skipping.")
	}
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

func checkIfCommandExists(command string, requiredMajorVersion int, requiredMinorVersion int) {
	path, err := exec.LookPath(command)
	if err != nil {
		log.Fatalln(command+" not found, please install. Error: ", err)
	}

	log.Println(command + " command located: " + path)

	version := exec.Command(command, "--version")

	versionOutput, versionError := version.Output()
	if versionError != nil {
		log.Fatalln("Unable to run " + command + " --version, please check your installation.")
	}

	versionString := string(versionOutput)
	buildRegex := regexp.MustCompile(`, .*\n`)
	versionWithoutBuild := buildRegex.ReplaceAllString(strings.ReplaceAll(versionString, "Docker version ", ""), ``)
	versionComponents := strings.Split(versionWithoutBuild, ".")
	majorVersion, _ := strconv.Atoi(versionComponents[0])
	minorVersion, _ := strconv.Atoi(versionComponents[1])
	if majorVersion < requiredMajorVersion || minorVersion < requiredMinorVersion {
		log.Fatalln("Current version of " + command + " is " + versionWithoutBuild + "." +
			" Min version required: " + strconv.Itoa(requiredMajorVersion) + "." + strconv.Itoa(requiredMinorVersion) + ".0")
	}
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
