package executor

import (
	"log"
	"os/exec"

	"github.com/jc/steakpie/internal/config"
)

// Runner executes a shell command, returns combined stdout+stderr and error.
type Runner interface {
	Run(cmd string) (output string, err error)
}

// ShellRunner runs commands via bash -lc (login shell for env vars).
type ShellRunner struct{}

func (s ShellRunner) Run(cmd string) (string, error) {
	out, err := exec.Command("bash", "-lc", cmd).CombinedOutput()
	return string(out), err
}

// Execute runs a list of commands for a webhook event.
// Top-level siblings run sequentially in parallel flow (continue on failure).
// Children only run if their parent succeeds.
func Execute(runner Runner, packageName, deliveryID string, commands []config.Command) {
	log.Printf("start webhook for %s received with id: %s", packageName, deliveryID)

	if len(commands) == 1 && len(commands[0].Children) > 0 {
		log.Printf("nested flow detected")
	} else if len(commands) > 1 {
		log.Printf("parallel flow detected")
	}

	executeLevel(runner, commands)

	log.Printf("end webhook for %s with id: %s", packageName, deliveryID)
}

// executeLevel runs a slice of sibling commands. Siblings continue even if one fails.
// Children of a command only run if the parent succeeds.
func executeLevel(runner Runner, commands []config.Command) {
	total := len(commands)
	for i, cmd := range commands {
		n := i + 1
		log.Printf("running command %d of %d: %s", n, total, cmd.Cmd)

		output, err := runner.Run(cmd.Cmd)
		if output != "" {
			log.Printf("output: %s", output)
		}

		if err != nil {
			log.Printf("command %d of %d failed: %v", n, total, err)
			continue
		}

		log.Printf("command %d of %d succeeded", n, total)

		if len(cmd.Children) > 0 {
			executeLevel(runner, cmd.Children)
		}
	}
}
