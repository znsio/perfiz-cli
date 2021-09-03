package configuration

type PerfizConfig struct {
	KarateFeaturesDir     string `yaml:"karateFeaturesDir"`
	KarateEnv             string `yaml:"karateEnv"`
	GatlingSimulationsDir string `yaml:"gatlingSimulationsDir"`
}

func GetGatlingSimulationsDir(workingDir string, config *PerfizConfig) string {
	if config.GatlingSimulationsDir == "" {
		return ""
	}
	return workingDir + "/" + config.GatlingSimulationsDir
}
