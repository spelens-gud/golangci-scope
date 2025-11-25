package build

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spelens-gud/logger"
)

// Run excutes the main package in addition with the internal goc features
func (b *Build) Run() error {
	cmd := exec.Command("/bin/bash", "-c", "go run "+b.BuildFlags+" "+b.GoRunExecFlag+" "+b.Packages+" "+b.GoRunArguments)
	cmd.Dir = b.TmpWorkingDir

	if b.NewGOPATH != "" {
		// Change to temp GOPATH for go install command
		cmd.Env = append(os.Environ(), fmt.Sprintf("GOPATH=%v", b.NewGOPATH))
	}

	logger.Infof("go build cmd is: %v", cmd.Args)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("fail to execute: %v, err: %w", cmd.Args, err)
	}

	if err = cmd.Wait(); err != nil {
		return fmt.Errorf("fail to execute: %v, err: %w", cmd.Args, err)
	}

	return nil
}
