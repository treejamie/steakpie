package executor

import (
	"log"
	"os/exec"

	"github.com/jc/steakpie/internal/config"
)

// Runner executes a shell command in a given directory, returns combined stdout+stderr and error.
type Runner interface {
	Run(cmd string, dir string) (output string, err error)
}

// ShellRunner runs commands via bash -lc (login shell for env vars).
type ShellRunner struct{}

func (s ShellRunner) Run(cmd string, dir string) (string, error) {
	c := exec.Command("bash", "-lc", cmd)
	if dir != "" {
		c.Dir = dir
	}
	out, err := c.CombinedOutput()
	return string(out), err
}

// Execute runs commands for a webhook event, grouped by directory.
// Each directory's commands run sequentially. Children only run if their parent succeeds.
func Execute(runner Runner, packageName, deliveryID string, dirCommands map[string][]config.Command) {
	log.Printf("start webhook for %s received with id: %s", packageName, deliveryID)

	for dir, commands := range dirCommands {
		log.Printf("executing in directory: %s", dir)
		executeLevel(runner, dir, commands)
	}

	log.Printf("end webhook for %s with id: %s", packageName, deliveryID)
}

// executeLevel runs a slice of sibling commands. Siblings continue even if one fails.
// Children of a command only run if the parent succeeds.
func executeLevel(runner Runner, dir string, commands []config.Command) {
	total := len(commands)
	for i, cmd := range commands {
		n := i + 1
		log.Printf("running command %d of %d: %s", n, total, cmd.Cmd)

		output, err := runner.Run(cmd.Cmd, dir)
		if output != "" {
			log.Printf("output: %s", output)
		}

		if err != nil {
			log.Printf("command %d of %d failed: %v", n, total, err)
			continue
		}

		log.Printf("command %d of %d succeeded", n, total)

		if len(cmd.Children) > 0 {
			executeLevel(runner, dir, cmd.Children)
		}
	}
}
