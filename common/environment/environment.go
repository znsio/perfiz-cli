package environment

import (
	"errors"
	cmd "github.com/znsio/perfiz-cli/common/command"
	"log"
	"os"
	"os/exec"
	"os/user"
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
	path, err := exec.LookPath(command)
	if err != nil {
		log.Fatalln(command+" not found, please install. Error: ", err)
	}

	log.Println(command + " command located: " + path)

	versionCommand := cmd.Create(command, "--version")
	versionCheckOkay, err := CheckCommandVersion(versionCommand, requiredMajorVersion, requiredMinorVersion)
	if !versionCheckOkay {
		log.Fatalln("Error locating " + command + ":" + err.Error())
	}
}

func CheckCommandVersion(version cmd.Command, requiredMajorVersion int, requiredMinorVersion int) (bool, error) {
	versionOutput, versionError := version.Execute()
	if versionError != nil {
		return false, versionError
	}

	versionString := string(versionOutput)

	commandRegex := regexp.MustCompile(`^.* version `)
	versionStringWithoutCommand := commandRegex.ReplaceAllString(versionString, ``)
	buildRegex := regexp.MustCompile(`, .*`)
	versionWithoutBuild := buildRegex.ReplaceAllString(strings.ReplaceAll(versionStringWithoutCommand, "Docker version ", ""), ``)
	releaseCandidateRegex := regexp.MustCompile(`-.*`)
	versionWithoutReleaseCandidate := releaseCandidateRegex.ReplaceAllString(strings.ReplaceAll(versionWithoutBuild, "v", ""), ``)
	versionComponents := strings.Split(versionWithoutReleaseCandidate, ".")
	majorVersion, _ := strconv.Atoi(versionComponents[0])
	minorVersion, _ := strconv.Atoi(versionComponents[1])

	if majorVersion > requiredMajorVersion {
		return true, nil
	}
	if majorVersion == requiredMajorVersion && minorVersion >= requiredMinorVersion {
		return true, nil
	}

	return false, errors.New("Current version is " + versionWithoutBuild + "." +
		" Min version required: " + strconv.Itoa(requiredMajorVersion) + "." + strconv.Itoa(requiredMinorVersion) + ".0")
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

func GetUserIdAndGroupId() (string, string) {
	current, err := user.Current()
	if err != nil {
		log.Println("Error getting current user: " + err.Error())
		log.Fatalln(err)
	}
	return current.Uid, current.Gid
}
