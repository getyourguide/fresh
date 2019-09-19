package runner

import (
	"io"
	"os/exec"
	"strconv"
)

func run() bool {
	runnerLog("Running...")

	cmd := runCmd()
	runnerLog("Runner command: %s %s", cmd.Path, cmd.Args)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		fatal(err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fatal(err)
	}

	err = cmd.Start()
	if err != nil {
		fatal(err)
	}

	var debugCmd = &exec.Cmd{}
	if debugEnabled() {
		debugCmd = exec.Command("dlv", "--listen=:"+debugPort(), "--headless=true", "--accept-multiclient", "--api-version=2", "--continue", "attach", strconv.Itoa(cmd.Process.Pid))
		err = debugCmd.Start()
		if err != nil {
			fatal(err)
		}
	}

	go io.Copy(appLogWriter{}, stderr)
	go io.Copy(appLogWriter{}, stdout)

	go func() {
		<-stopChannel
		pid := cmd.Process.Pid
		runnerLog("Killing PID %d", pid)
		cmd.Process.Kill()
		if debugCmd.Process.Pid != 0 { // Not null
			runnerLog("Killing debug server PID %d", debugCmd.Process.Pid)
			debugCmd.Process.Kill()
		}
	}()

	return true
}
