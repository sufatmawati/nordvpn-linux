package childprocess

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"google.golang.org/grpc/status"
)

type ProcessClient interface {
	Ping(nowait bool) error
	Stop() error
}

type GRPCChildProcessManager struct {
	processClient     ProcessClient
	processBinaryPath string
}

func NewGRPCChildProcessManager(processClient ProcessClient, processBinaryPath string) *GRPCChildProcessManager {
	return &GRPCChildProcessManager{
		processClient:     processClient,
		processBinaryPath: processBinaryPath,
	}
}

func (g *GRPCChildProcessManager) StartProcess() (StartupErrorCode, error) {
	errChan := make(chan error)
	go func() {
		// #nosec G204 -- arg values are known before even running the program
		_, err := exec.Command(g.processBinaryPath).Output()
		errChan <- err
	}()

	pingChan := make(chan error)
	// Start another goroutine where we ping the WaitForReady option, so that server has time to start up before we run
	// the acctual command.
	go func() {
		err := g.processClient.Ping(false)
		pingChan <- err
	}()

	select {
	case err := <-errChan:
		if err == nil {
			return 0, fmt.Errorf("process finished unexpectedly")
		}
		var exiterr *exec.ExitError
		if errors.As(err, &exiterr) {
			exitCode := StartupErrorCode(exiterr.ExitCode())
			return exitCode, nil
		}
		return 0, fmt.Errorf("failed to start the process: %w", err)
	case err := <-pingChan:
		if err != nil {
			return 0, fmt.Errorf("failed to ping the process after starting: %w", err)
		}

		// Process was started and pinged successfully.
		return 0, nil
	}
}

func (g *GRPCChildProcessManager) StopProcess() error {
	err := g.processClient.Stop()
	if err != nil {
		return fmt.Errorf("stopping fileshare client: %w", err)
	}

	return nil
}

func (g *GRPCChildProcessManager) ProcessStatus() ProcessStatus {
	err := g.processClient.Ping(true)
	if err != nil {
		if strings.Contains(status.Convert(err).Message(), "permission denied") {
			return RunningForOtherUser
		}
		return NotRunning
	}

	return Running
}