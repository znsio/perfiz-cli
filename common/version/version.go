package version

import (
	"github.com/znsio/perfiz-cli/common/constants"
	"github.com/znsio/perfiz-cli/common/environment"
	"io/ioutil"
	"log"
)

func GetPerfizVersion() string {
	perfizHome := environment.GetEnvVariable(constants.PERFIZ_HOME_ENV_VARIABLE)
	perfizVersion, perfizVersionErr := ioutil.ReadFile(perfizHome + "/.VERSION")
	if perfizVersionErr != nil {
		log.Println("Unable to read Perfiz Version File: " + perfizHome + "/.VERSION")
	}
	return string(perfizVersion)
}