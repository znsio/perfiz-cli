package constants

const (
	PERFIZ_HOME_ENV_VARIABLE        = "PERFIZ_HOME"
	DEFAULT_CONFIG_FILE             = "perfiz.yml"
	PERFIZ_FOLDER                   = "./perfiz"
	GATLING_CONF                    = "gatling.conf"
	GATLING_CONF_PATH               = PERFIZ_FOLDER + "/gatling/"
	GATLING_RESULTS_DIR             = "perfiz/gatling_data/results"
	GRAFANA_DASHBOARDS_DIRECTORY    = PERFIZ_FOLDER + "/dashboards"
	PROMETHEUS_CONFIG_DIR           = PERFIZ_FOLDER + "/prometheus"
	PROMETHEUS_CONFIG               = PROMETHEUS_CONFIG_DIR + "/prometheus.yml"
	DOCKER_COMPOSE_ENV_FILE         = "/.env"
	DOCKER_MAJOR_VERSION            = 20
	DOCKER_MINOR_VERSION            = 10
	DOCKER_COMPOSE_MAJOR_VERSION    = 1
	DOCKER_COMPOSE_MINOR_VERSION    = 29
	PERFIZ_CLI_VERSION              = "0.0.20"
	PERFIZ_GATLING_SIMULATION_CLASS = "org.znsio.perfiz.PerfizSimulation"

	SKIP_TEMPLATE_MESSAGE = " is already present. Skipping."
)
