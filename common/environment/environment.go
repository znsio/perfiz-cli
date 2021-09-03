package environment

import (
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

func GetEnvVariable(envVariableName string) string {
	envVariable := os.Getenv(envVariableName)
	if len(envVariable) == 0 {
		log.Fatalln("Please set " + envVariableName + " environment variable")
	}
	log.Println(envVariableName + ": " + envVariable)
	return envVariable
}

func CheckIfCommandExists(command string, requiredMajorVersion int, requiredMinorVersion int) {
	versionString := GetCommandVersion(command)
	commandRegex := regexp.MustCompile(`^.* version `)
	versionStringWithoutCommand := commandRegex.ReplaceAllString(versionString, ``)
	buildRegex := regexp.MustCompile(`, .*\n`)
	versionWithoutBuild := buildRegex.ReplaceAllString(strings.ReplaceAll(versionStringWithoutCommand, "Docker version ", ""), ``)
	versionComponents := strings.Split(versionWithoutBuild, ".")
	majorVersion, _ := strconv.Atoi(versionComponents[0])
	minorVersion, _ := strconv.Atoi(versionComponents[1])
	if majorVersion < requiredMajorVersion || minorVersion < requiredMinorVersion {
		log.Fatalln("Current version of " + command + " is " + versionWithoutBuild + "." +
			" Min version required: " + strconv.Itoa(requiredMajorVersion) + "." + strconv.Itoa(requiredMinorVersion) + ".0")
	}
}

func GetCommandVersion(command string) string {
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
	return versionString
}