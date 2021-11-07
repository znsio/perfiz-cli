package cmd

import (
	"bufio"
	"errors"
	"github.com/otiai10/copy"
	"github.com/spf13/cobra"
	"github.com/znsio/perfiz-cli/common/configuration"
	"github.com/znsio/perfiz-cli/common/constants"
	env "github.com/znsio/perfiz-cli/common/environment"
	"github.com/znsio/perfiz-cli/common/path"
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

func init() {
	rootCmd.AddCommand(cmdTest)
}

var cmdTest = &cobra.Command{
	Use:   "test [perfiz config file name]",
	Short: "Run Gatling Performance Test",
	Long:  `Run Gatling Performance Tests as per the configuration in perfiz.yml`,
	Args: func(cmd *cobra.Command, args []string) error {
		_, perfizYmlErr := os.Open(constants.DEFAULT_CONFIG_FILE)
		if len(args) < 1 {
			if perfizYmlErr != nil {
				return errors.New("Default Config: " + constants.DEFAULT_CONFIG_FILE + " not found. Please create " + constants.DEFAULT_CONFIG_FILE + " or provide name of config file as argument. Please see https://github.com/znsio/perfiz for instructions and / or run 'init' command and perfiz will add a config file template to help you get started.")
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
		perfizHome := env.GetEnvVariable(constants.PERFIZ_HOME_ENV_VARIABLE)
		env.CheckIfCommandExists("docker", constants.DOCKER_MAJOR_VERSION, constants.DOCKER_MINOR_VERSION)
		env.CheckIfCommandExists("docker-compose", constants.DOCKER_COMPOSE_MAJOR_VERSION, constants.DOCKER_COMPOSE_MINOR_VERSION)

		var configFile string
		if len(args) == 1 {
			configFile = args[0]
		} else {
			configFile = constants.DEFAULT_CONFIG_FILE
		}
		log.Println("Perfiz Config File: " + configFile)
		perfizConfig := &configuration.PerfizConfig{}
		b, _ := ioutil.ReadFile(configFile)
		configParseError := yaml.Unmarshal(b, perfizConfig)
		if configParseError != nil {
			log.Fatal(configParseError)
		}
		karateFeaturesDir := workingDir + "/" + perfizConfig.KarateFeaturesDir
		if !path.IsDir(karateFeaturesDir) {
			log.Fatalln("Configuration error in perfiz.yml. karateFeaturesDir: " + perfizConfig.KarateFeaturesDir + ". " + karateFeaturesDir + " is not a directory. Please note that karateFeaturesDir has to be relative to perfiz.yml location.")
		}
		gatlingSimulationsDir := configuration.GetGatlingSimulationsDir(workingDir, perfizConfig)
		if gatlingSimulationsDir != "" && !path.IsDir(gatlingSimulationsDir) {
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
					return !path.IsDir(src) && !strings.HasSuffix(src, ".scala"), nil
				},
			}
			copy.Copy(gatlingSimulationsDir, perfizHome+"/src/test/scala", onlyScalaSimulationFiles)
		}

		_, gatlingConfErr := os.Open(constants.GATLING_CONF_PATH + constants.GATLING_CONF)
		if gatlingConfErr == nil {
			log.Println("Copying Gatling Configuration " + constants.GATLING_CONF_PATH + constants.GATLING_CONF)
			copy.Copy(constants.GATLING_CONF_PATH+constants.GATLING_CONF, perfizHome+"/src/test/resources/"+constants.GATLING_CONF)
		}

		perfizMavenRepo := perfizHome + "/.m2"
		if path.IsDir(perfizMavenRepo) {
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

		uid, gid := env.GetUserIdAndGroupId()

		dockerCommandArguments := []string{"run", "--rm", "--name", "perfiz-gatling",
			"-v", perfizMavenRepo + ":/var/maven/.m2",
			"-v", perfizHome + ":/var/maven",
			"-v", workingDir + "/" + constants.GATLING_RESULTS_DIR + ":/usr/src/performance-testing/results",
			"-v", perfizHome + ":/usr/src/performance-testing",
			"-v", karateFeaturesDir + ":/usr/src/karate-features",
			"-v", workingDir + "/" + configFile + ":/usr/src/perfiz.yml",
			"-e", "KARATE_FEATURES=/usr/src/karate-features",
			"-e", "MAVEN_CONFIG=/var/maven/.m2",
			"-w", "/usr/src/performance-testing",
			"--user", uid + ":" + gid,
			"--network", "perfiz-network",
			"maven:3.8-jdk-8", "mvn", "clean", "test-compile", "gatling:test", "-DPERFIZ=/usr/src/perfiz.yml", "-Duser.home=/var/maven"}

		karateEnv := perfizConfig.KarateEnv
		if karateEnv != "" {
			log.Println("Setting karate.env to " + karateEnv)
			dockerCommandArguments = append(dockerCommandArguments, "-Dkarate.env="+karateEnv)
		}

		gatlingSimulationClass := perfizConfig.GatlingSimulationClass
		if gatlingSimulationClass != "" {
			log.Println("Setting gatling.simulationClass to " + gatlingSimulationClass)
			dockerCommandArguments = append(dockerCommandArguments, "-Dgatling.simulationClass="+gatlingSimulationClass)
		} else {
			log.Println("Setting gatling.simulationClass to " + constants.PERFIZ_GATLING_SIMULATION_CLASS)
			dockerCommandArguments = append(dockerCommandArguments, "-Dgatling.simulationClass="+constants.PERFIZ_GATLING_SIMULATION_CLASS)
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

func logStreamingOutput(output io.ReadCloser) {
	outputScanner := bufio.NewScanner(output)
	outputScanner.Split(bufio.ScanLines)
	for outputScanner.Scan() {
		log.Println(outputScanner.Text())
	}
}
