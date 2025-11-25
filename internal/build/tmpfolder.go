package build

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/spelens-gud/golangci-scope/internal/cover"
	"github.com/spelens-gud/logger"
	"github.com/spf13/viper"
)

// MvProjectsToTmp moves the projects into a temporary directory
func (b *Build) MvProjectsToTmp() error {
	listArgs := []string{"-json"}
	if len(b.BuildFlags) != 0 {
		listArgs = append(listArgs, b.BuildFlags)
	}
	listArgs = append(listArgs, "./...")
	var err error
	b.Pkgs, err = cover.ListPackages(b.WorkingDir, strings.Join(listArgs, " "), "")
	if err != nil {
		logger.Error(err.Error())
		return err
	}

	err = b.mvProjectsToTmp()
	if err != nil {
		logger.Errorf("Fail to move the project to temporary directory")
		return err
	}
	b.OriGOPATH = os.Getenv("GOPATH")
	if b.IsMod {
		b.NewGOPATH = ""
	} else if b.OriGOPATH == "" {
		b.NewGOPATH = b.TmpDir
	} else {
		b.NewGOPATH = fmt.Sprintf("%v:%v", b.TmpDir, b.OriGOPATH)
	}
	// fix #14: unable to build project not in GOPATH in legacy mode
	// this kind of project does not have a pkg.Root value
	// go 1.11, 1.12 has no pkg.Root,
	// so add b.IsMod == false as secondary judgement
	if b.NewGOPATH == "" && b.Root == "" && !b.IsMod {
		b.NewGOPATH = b.OriGOPATH
	}
	logger.Infof("New GOPATH: %v", b.NewGOPATH)
	return nil
}

func (b *Build) mvProjectsToTmp() error {
	b.TmpDir = filepath.Join(os.TempDir(), tmpFolderName(b.WorkingDir))

	// Delete previous tmp folder and its content
	os.RemoveAll(b.TmpDir)
	// Create a new tmp folder and a new importpath for storing cover variables
	b.GlobalCoverVarImportPath = filepath.Join("src", tmpPackageName(b.WorkingDir))
	err := os.MkdirAll(filepath.Join(b.TmpDir, b.GlobalCoverVarImportPath), os.ModePerm)
	if err != nil {
		return fmt.Errorf("Fail to create the temporary build directory. The err is: %v", err)
	}
	logger.Infof("Tmp project generated in: %v", b.TmpDir)

	// traverse pkg list to get project meta info
	b.IsMod, b.Root, err = b.traversePkgsList()
	logger.Infof("mod project? %v", b.IsMod)
	if errors.Is(err, ErrShouldNotReached) {
		return fmt.Errorf("mvProjectsToTmp with a empty project: %w", err)
	}
	// we should get corresponding working directory in temporary directory
	b.TmpWorkingDir, err = b.getTmpwd()
	if err != nil {
		return fmt.Errorf("getTmpwd failed with error: %w", err)
	}
	// issue #14
	// if b.Root == "", then the project is non-standard project
	// known cases:
	// 1. a legacy project, but not in any GOPATH, will cause the b.Root == ""
	if b.IsMod == false && b.Root != "" {
		b.cpLegacyProject()
	} else if b.IsMod == true { // go 1.11, 1.12 has no Build.Root
		b.cpGoModulesProject()
		updated, newGoModContent, err := b.updateGoModFile()
		if err != nil {
			return fmt.Errorf("fail to generate new go.mod: %v", err)
		}
		if updated {
			logger.Info("go.mod needs rewrite")
			tmpModFile := filepath.Join(b.TmpDir, "go.mod")
			err := ioutil.WriteFile(tmpModFile, newGoModContent, os.ModePerm)
			if err != nil {
				return fmt.Errorf("fail to update go.mod: %v", err)
			}
		}
	} else if b.IsMod == false && b.Root == "" {
		b.TmpWorkingDir = b.TmpDir
		b.cpNonStandardLegacy()
	}

	logger.Infof("New workingdir in tmp directory in: %v", b.TmpWorkingDir)
	return nil
}

func tmpFolderName(path string) string {
	sum := sha256.Sum256([]byte(path))
	h := fmt.Sprintf("%x", sum[:6])

	return "goc-build-" + h
}
func tmpPackageName(path string) string {
	sum := sha256.Sum256([]byte(path))
	h := fmt.Sprintf("%x", sum[:6])

	return "gocbuild" + h
}
func (b *Build) traversePkgsList() (isMod bool, root string, err error) {
	for _, v := range b.Pkgs {
		// get root
		root = v.Root
		if v.Module == nil {
			return
		}
		isMod = true
		b.ModRoot = v.Module.Dir
		b.ModRootPath = v.Module.Path
		return
	}
	logger.Error(ErrShouldNotReached.Error())
	err = ErrShouldNotReached
	return
}

func (b *Build) getTmpwd() (string, error) {
	for _, pkg := range b.Pkgs {
		var index int
		var parentPath string
		if b.IsMod == false {
			index = strings.Index(b.WorkingDir, pkg.Root)
			parentPath = pkg.Root
		} else {
			index = strings.Index(b.WorkingDir, pkg.Module.Dir)
			parentPath = pkg.Module.Dir
		}

		if index == -1 {
			return "", ErrGocShouldExecInProject
		}
		// b.TmpWorkingDir = filepath.Join(b.TmpDir, path[len(parentPath):])
		return filepath.Join(b.TmpDir, b.WorkingDir[len(parentPath):]), nil
	}

	return "", ErrShouldNotReached
}
func (b *Build) Clean() error {
	if !viper.GetBool("debug") {
		return os.RemoveAll(b.TmpDir)
	}
	return nil
}
