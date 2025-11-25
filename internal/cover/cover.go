package cover

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	tool "github.com/spelens-gud/golangci-scope/internal/cover/internal"
	"github.com/spelens-gud/logger"
)

var (
	// ErrCoverPkgFailed represents the error that fails to inject the package
	ErrCoverPkgFailed = errors.New("fail to inject code to project")
	// ErrCoverListFailed represents the error that fails to list package dependencies
	ErrCoverListFailed = errors.New("fail to list package dependencies")
)

type TestCover struct {
	Mode                     string
	AgentPort                string
	Center                   string // cover profile host center
	Singleton                bool
	MainPkgCover             *PackageCover
	DepsCover                []*PackageCover
	CacheCover               map[string]*PackageCover
	GlobalCoverVarImportPath string
}
type PackageCover struct {
	Package *Package
	Vars    map[string]*FileVar
}
type FileVar struct {
	File string
	Var  string
}
type Package struct {
	Dir        string `json:"Dir"`        // directory containing package sources
	ImportPath string `json:"ImportPath"` // import path of package in dir
	Name       string `json:"Name"`       // package name
	Target     string `json:",omitempty"` // installed target for this package (may be executable)
	Root       string `json:",omitempty"` // Go root, Go path dir, or module root dir containing this package

	Module   *ModulePublic `json:",omitempty"`         // info about package's module, if any
	Goroot   bool          `json:"Goroot,omitempty"`   // is this package in the Go root?
	Standard bool          `json:"Standard,omitempty"` // is this package part of the standard Go library?
	DepOnly  bool          `json:"DepOnly,omitempty"`  // package is only a dependency, not explicitly listed

	// Source files
	GoFiles  []string `json:"GoFiles,omitempty"`  // .go source files (excluding CgoFiles, TestGoFiles, XTestGoFiles)
	CgoFiles []string `json:"CgoFiles,omitempty"` // .go source files that import "C"

	// Dependency information
	Deps      []string          `json:"Deps,omitempty"` // all (recursively) imported dependencies
	Imports   []string          `json:",omitempty"`     // import paths used by this package
	ImportMap map[string]string `json:",omitempty"`     // map from source import to ImportPath (identity entries omitted)

	// Error information
	Incomplete bool            `json:"Incomplete,omitempty"` // this package or a dependency has an error
	Error      *PackageError   `json:"Error,omitempty"`      // error loading package
	DepsErrors []*PackageError `json:"DepsErrors,omitempty"` // errors loading dependencies
}
type ModulePublic struct {
	Path      string        `json:",omitempty"` // module path
	Version   string        `json:",omitempty"` // module version
	Versions  []string      `json:",omitempty"` // available module versions
	Replace   *ModulePublic `json:",omitempty"` // replaced by this module
	Time      *time.Time    `json:",omitempty"` // time version was created
	Update    *ModulePublic `json:",omitempty"` // available update (with -u)
	Main      bool          `json:",omitempty"` // is this the main module?
	Indirect  bool          `json:",omitempty"` // module is only indirectly needed by main module
	Dir       string        `json:",omitempty"` // directory holding local copy of files, if any
	GoMod     string        `json:",omitempty"` // path to go.mod file describing module, if any
	GoVersion string        `json:",omitempty"` // go version used in module
	Error     *ModuleError  `json:",omitempty"` // error loading module
}
type ModuleError struct {
	Err string // error text
}
type PackageError struct {
	ImportStack []string // shortest path from package named on command line to this one
	Pos         string   // position of error (if present, file:line:col)
	Err         string   // the error itself
}

func ListPackages(dir string, args string, newgopath string) (map[string]*Package, error) {
	cmd := exec.Command("/bin/bash", "-c", "go list "+args)
	log.Printf("go list cmd is: %v", cmd.Args)
	cmd.Dir = dir
	if newgopath != "" {
		cmd.Env = append(os.Environ(), fmt.Sprintf("GOPATH=%v", newgopath))
	}
	var errbuf bytes.Buffer
	cmd.Stderr = &errbuf
	out, err := cmd.Output()
	if err != nil {
		logger.Errorf("excute `go list -json ./...` command failed, err: %v, stdout: %v, stderr: %v", err, string(out), errbuf.String())
		return nil, ErrCoverListFailed
	}
	logger.Infof("\n%v", errbuf.String())
	dec := json.NewDecoder(bytes.NewReader(out))
	pkgs := make(map[string]*Package, 0)
	for {
		var pkg Package
		if err := dec.Decode(&pkg); err != nil {
			if err == io.EOF {
				break
			}
			logger.Errorf("reading go list output: %v", err)
			return nil, ErrCoverListFailed
		}
		if pkg.Error != nil {
			logger.Errorf("list package %s failed with output: %v", pkg.ImportPath, pkg.Error)
			return nil, ErrCoverPkgFailed
		}

		// for _, err := range pkg.DepsErrors {
		// 	log.Fatalf("dependency package list failed, err: %v", err)
		// }

		pkgs[pkg.ImportPath] = &pkg
	}
	return pkgs, nil
}

type CoverInfo struct {
	Target                   string
	GoPath                   string
	IsMod                    bool
	ModRootPath              string
	GlobalCoverVarImportPath string // path for the injected global cover var file
	OneMainPackage           bool
	Args                     string
	Mode                     string
	AgentPort                string
	Center                   string
	Singleton                bool
}

func Execute(coverInfo *CoverInfo) error {
	target := coverInfo.Target
	newGopath := coverInfo.GoPath
	// oneMainPackage := coverInfo.OneMainPackage
	args := coverInfo.Args
	mode := coverInfo.Mode
	agentPort := coverInfo.AgentPort
	center := coverInfo.Center
	singleton := coverInfo.Singleton
	globalCoverVarImportPath := coverInfo.GlobalCoverVarImportPath

	if coverInfo.IsMod {
		globalCoverVarImportPath = filepath.Join(coverInfo.ModRootPath, globalCoverVarImportPath)
	} else {
		globalCoverVarImportPath = filepath.Base(globalCoverVarImportPath)
	}

	if !isDirExist(target) {
		logger.Errorf("Target directory %s not exist", target)
		return ErrCoverPkgFailed
	}
	listArgs := []string{"-json"}
	if len(args) != 0 {
		listArgs = append(listArgs, args)
	}
	listArgs = append(listArgs, "./...")
	pkgs, err := ListPackages(target, strings.Join(listArgs, " "), newGopath)
	if err != nil {
		logger.Errorf("Fail to list all packages, the error: %v", err)
		return err
	}

	var seen = make(map[string]*PackageCover)
	// var seenCache = make(map[string]*PackageCover)
	allDecl := ""
	for _, pkg := range pkgs {
		if pkg.Name == "main" {
			logger.Infof("handle package: %v", pkg.ImportPath)
			// inject the main package
			mainCover, mainDecl := AddCounters(pkg, mode, globalCoverVarImportPath)
			allDecl += mainDecl
			// new a testcover for this service
			tc := TestCover{
				Mode:                     mode,
				AgentPort:                agentPort,
				Center:                   center,
				Singleton:                singleton,
				MainPkgCover:             mainCover,
				GlobalCoverVarImportPath: globalCoverVarImportPath,
			}

			// handle its dependency
			// var internalPkgCache = make(map[string][]*PackageCover)
			tc.CacheCover = make(map[string]*PackageCover)
			for _, dep := range pkg.Deps {
				if packageCover, ok := seen[dep]; ok {
					tc.DepsCover = append(tc.DepsCover, packageCover)
					continue
				}

				//only focus package neither standard Go library nor dependency library
				if depPkg, ok := pkgs[dep]; ok {
					packageCover, depDecl := AddCounters(depPkg, mode, globalCoverVarImportPath)
					allDecl += depDecl
					tc.DepsCover = append(tc.DepsCover, packageCover)
					seen[dep] = packageCover
				}
			}

			// inject Http Cover APIs
			var httpCoverApis = fmt.Sprintf("%s/http_cover_apis_auto_generated.go", pkg.Dir)
			if err := InjectCountersHandlers(tc, httpCoverApis); err != nil {
				logger.Errorf("failed to inject counters for package: %s, err: %v", pkg.ImportPath, err)
				return ErrCoverPkgFailed
			}
		}
	}

	return injectGlobalCoverVarFile(coverInfo, allDecl)
}
func isDirExist(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

func AddCounters(pkg *Package, mode string, globalCoverVarImportPath string) (*PackageCover, string) {
	coverVarMap := declareCoverVars(pkg)

	decl := ""
	for file, coverVar := range coverVarMap {
		decl += "\n" + tool.Annotate(path.Join(pkg.Dir, file), mode, coverVar.Var, globalCoverVarImportPath) + "\n"
	}

	return &PackageCover{
		Package: pkg,
		Vars:    coverVarMap,
	}, decl
}
func declareCoverVars(p *Package) map[string]*FileVar {
	coverVars := make(map[string]*FileVar)
	coverIndex := 0
	// We create the cover counters as new top-level variables in the package.
	// We need to avoid collisions with user variables (GoCover_0 is unlikely but still)
	// and more importantly with dot imports of other covered packages,
	// so we append 12 hex digits from the SHA-256 of the import path.
	// The point is only to avoid accidents, not to defeat users determined to
	// break things.
	sum := sha256.Sum256([]byte(p.ImportPath))
	h := fmt.Sprintf("%x", sum[:6])
	for _, file := range p.GoFiles {
		// These names appear in the cmd/cover HTML interface.
		var longFile = path.Join(p.ImportPath, file)
		coverVars[file] = &FileVar{
			File: longFile,
			Var:  fmt.Sprintf("GoCover_%d_%x", coverIndex, h),
		}
		coverIndex++
	}

	for _, file := range p.CgoFiles {
		// These names appear in the cmd/cover HTML interface.
		var longFile = path.Join(p.ImportPath, file)
		coverVars[file] = &FileVar{
			File: longFile,
			Var:  fmt.Sprintf("GoCover_%d_%x", coverIndex, h),
		}
		coverIndex++
	}

	return coverVars
}
