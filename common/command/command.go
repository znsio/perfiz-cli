package command

import "os/exec"

type Command interface {
	Execute() ([]byte, error)
}

type OsExecCommand struct {
	Name string
	Args []string
}

func Create(name string, arg ...string) Command {
	var command Command = &OsExecCommand{
		name,
		append([]string{}, arg...),
	}
	return command
}

func (cmd *OsExecCommand) Execute() ([]byte, error) {
	command := exec.Command(cmd.Name, cmd.Args...)
	return command.Output()
}
