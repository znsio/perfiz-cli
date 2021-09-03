package cmd

import (
	"github.com/otiai10/copy"
	"github.com/spf13/cobra"
	"github.com/znsio/perfiz-cli/common/constants"
	env "github.com/znsio/perfiz-cli/common/environment"
	"github.com/znsio/perfiz-cli/common/path"
	"log"
	"os"
)

func init() {
	rootCmd.AddCommand(cmdInit)
}

var cmdInit = &cobra.Command{
	Use:   "init",
	Short: "Add Perfiz Config Templates and Dirs",
	Long: `Add Perfiz Config YML template, Directories for Grafana Dashboards,
                Prometheus Configs and update .gitignore`,
	Args: cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("Staring Init")
		perfizHome := env.GetEnvVariable(constants.PERFIZ_HOME_ENV_VARIABLE)
		addTemplateIfMissing(perfizHome, constants.DEFAULT_CONFIG_FILE, "./")
		addTemplateIfMissing(perfizHome, constants.GATLING_CONF, constants.GATLING_CONF_PATH)
		if !path.IsDir(constants.GRAFANA_DASHBOARDS_DIRECTORY) {
			log.Println("Creating Grafana Dashboard dir " + constants.GRAFANA_DASHBOARDS_DIRECTORY + ". Add Grafana Dashboard JSONs here.")
			os.MkdirAll(constants.GRAFANA_DASHBOARDS_DIRECTORY, 0755)
			log.Println("Adding sample dashboard json " + constants.GRAFANA_DASHBOARDS_DIRECTORY + "/dashboard.json as a reference to get you started.")
			copy.Copy(perfizHome+"/templates/dashboard.json", constants.GRAFANA_DASHBOARDS_DIRECTORY+"/dashboard.json")
		} else {
			log.Println(constants.GRAFANA_DASHBOARDS_DIRECTORY + " is already present. Skipping.")
		}
		_, prometheusConfigErr := os.Open(constants.PROMETHEUS_CONFIG)
		if prometheusConfigErr != nil {
			log.Println("Creating prometheus.yml template in " + constants.PROMETHEUS_CONFIG + ". Add scrape configs to this file.")
			os.MkdirAll(constants.PROMETHEUS_CONFIG_DIR, 0755)
			copy.Copy(perfizHome+"/templates/prometheus.yml", constants.PROMETHEUS_CONFIG)
		} else {
			log.Println(constants.PROMETHEUS_CONFIG + " is already present. Skipping.")
		}
		log.Println("Setting ./perfiz permissions to 0777 to allow Docker containers to access its contents")
		os.Chmod(constants.PERFIZ_FOLDER, 0777)
		log.Println("Init Completed")
		log.Println("Please add below line to your .gitignore to avoid checking in Prometheus and Grafana Data to version control")
		log.Println("perfiz/*_data")
	},
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