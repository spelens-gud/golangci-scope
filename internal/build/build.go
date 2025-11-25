package build

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spelens-gud/golangci-scope/internal/cover"
	"github.com/spelens-gud/logger"
)

// Build is to describe the building/installing process of a goc build/install
type Build struct {
	Pkgs          map[string]*cover.Package // Pkg list parsed from "go list -json ./..." command
	NewGOPATH     string                    // the new GOPATH
	OriGOPATH     string                    // the original GOPATH
	WorkingDir    string                    // the working directory
	TmpDir        string                    // the temporary directory to build the project
	TmpWorkingDir string                    // the working directory in the temporary directory, which is corresponding to the current directory in the project directory
	IsMod         bool                      // determine whether it is a Mod project
	Root          string
	// go 1.11, go 1.12 has no Root
	// Project Root:
	// 1. legacy, root == GOPATH
	// 2. mod, root == go.mod Dir
	ModRoot     string // path for go.mod
	ModRootPath string // import path for the whole project
	Target      string // the binary name that go build generate
	// keep compatible with go commands:
	// go run [build flags] [-exec xprog] package [arguments...]
	// go build [-o output] [-i] [build flags] [packages]
	// go install [-i] [build flags] [packages]
	BuildFlags     string // Build flags
	Packages       string // Packages that needs to build
	GoRunExecFlag  string // for the -exec flags in go run command
	GoRunArguments string // for the '[arguments]' parameters in go run command

	OneMainPackage           bool   // whether this build is a go build or go install? true: build, false: install
	GlobalCoverVarImportPath string // Importpath for storing cover variables
	GlobalCoverVarFilePath   string // Importpath for storing cover variables
}

func NewBuild(buildflags string, args []string, workingDir string, outputDir string) (*Build, error) {
	if err := checkParameters(args, workingDir); err != nil {
		return nil, err
	}
	// buildflags = buildflags + " -o " + outputDir
	b := &Build{
		BuildFlags: buildflags,
		Packages:   strings.Join(args, " "),
		WorkingDir: workingDir,
	}
	if false == b.validatePackageForBuild() {
		logger.Error(ErrWrongPackageTypeForBuild.Error())
		return nil, ErrWrongPackageTypeForBuild
	}
	if err := b.MvProjectsToTmp(); err != nil {
		return nil, err
	}
	dir, err := b.determineOutputDir(outputDir)
	b.Target = dir
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (b *Build) determineOutputDir(outputDir string) (string, error) {
	if b.TmpDir == "" {
		return "", fmt.Errorf("can only be called after Build.MvProjectsToTmp(): %w", ErrEmptyTempWorkingDir)
	}

	// fix #43
	if outputDir != "" {
		abs, err := filepath.Abs(outputDir)
		if err != nil {
			return "", fmt.Errorf("Fail to transform the path: %v to absolute path: %v", outputDir, err)

		}
		return abs, nil
	}
	// fix #43
	// use target name from `go list -json ./...` of the main module
	targetName := ""
	for _, pkg := range b.Pkgs {
		if pkg.Name == "main" {
			if pkg.Target != "" {
				targetName = filepath.Base(pkg.Target)
			} else {
				targetName = filepath.Base(pkg.Dir)
			}
			break
		}
	}

	return filepath.Join(b.WorkingDir, targetName), nil
}

// validatePackageForBuild only allow . as package name
func (b *Build) validatePackageForBuild() bool {
	if b.Packages == "." || b.Packages == "" {
		return true
	}
	return false
}
func checkParameters(args []string, workingDir string) error {
	if len(args) > 1 {
		logger.Error(ErrTooManyArgs.Error())
		return ErrTooManyArgs
	}

	if workingDir == "" {
		return ErrInvalidWorkingDir
	}

	logger.Infof("Working directory: %v", workingDir)
	return nil
}
