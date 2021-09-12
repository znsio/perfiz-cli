package environment

import (
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

type CommandMock struct {
	mock.Mock
}

func (cmdMock *CommandMock) Execute() ([]byte, error) {
	args := cmdMock.Called()
	return []byte(args.String(0)), args.Error(1)
}

func Test_CheckCommandVersion_ReturnsTrueWhenCommandExistsWithVersionAtLeastEqualToMinRequirements(t *testing.T) {
	cmdMock := new(CommandMock)
	cmdMock.On("Execute").Return("Docker version 20.10.8, build 3967b7d", nil)
	commandExists, _ := CheckCommandVersion(cmdMock, 20, 10)
	assert.True(t, commandExists)
}

func Test_CheckCommandVersion_ReturnsFalseWhenCommandExistsWithMinorVersionLowerThanMinRequirements(t *testing.T) {
	cmdMock := new(CommandMock)
	cmdMock.On("Execute").Return("Docker version 20.9.8, build 3967b7d", nil)
	commandExists, error := CheckCommandVersion(cmdMock, 20, 10)
	assert.False(t, commandExists)
	assert.Equal(t, "Current version is 20.9.8. Min version required: 20.10.0", error.Error())
}

func Test_CheckCommandVersion_ReturnsFalseWhenCommandExistsWithMajorVersionLowerThanMinRequirements(t *testing.T) {
	cmdMock := new(CommandMock)
	cmdMock.On("Execute").Return("Docker version 19.11.8, build 3967b7d", nil)
	commandExists, error := CheckCommandVersion(cmdMock, 20, 10)
	assert.False(t, commandExists)
	assert.Equal(t, "Current version is 19.11.8. Min version required: 20.10.0", error.Error())
}

func Test_CheckCommandVersion_ParsesVersionStringsThatContainAlphabetsAndReleaseCandidateNumbers(t *testing.T) {
	cmdMock := new(CommandMock)
	cmdMock.On("Execute").Return("Docker Compose version v2.0.0-rc.2", nil)
	commandExists, error := CheckCommandVersion(cmdMock, 1, 29)
	assert.True(t, commandExists)
	assert.Nil(t, error)
}

func Test_CheckCommandVersion_ReturnsFalseWhenThereIsAnErrorRunningVersionCommand(t *testing.T) {
	cmdMock := new(CommandMock)
	cmdMock.On("Execute").Return("", errors.New("Error running version command"))
	commandExists, error := CheckCommandVersion(cmdMock, 1, 29)
	assert.False(t, commandExists)
	assert.Equal(t, "Error running version command", error.Error())
}
