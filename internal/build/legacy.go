package build

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spelens-gud/golangci-scope/internal/cover"
	"github.com/spelens-gud/logger"
	"github.com/tongjingran/copy"
)

func (b *Build) cpLegacyProject() {
	visited := make(map[string]bool)
	for k, v := range b.Pkgs {
		dst := filepath.Join(b.TmpDir, "src", k)
		src := v.Dir

		if _, ok := visited[src]; ok {
			// Skip if already copied
			continue
		}

		if err := copy.Copy(src, dst, copy.Options{Skip: skipCopy}); err != nil {
			logger.Errorf("Failed to Copy the folder from %v to %v, the error is: %v ", src, dst, err)
		}

		visited[src] = true

		b.cpDepPackages(v, visited)
	}
}
func (b *Build) cpDepPackages(pkg *cover.Package, visited map[string]bool) {
	gopath := pkg.Root
	for _, dep := range pkg.Deps {
		src := filepath.Join(gopath, "src", dep)
		// Check if copied
		if _, ok := visited[src]; ok {
			// Skip if already copied
			continue
		}
		// Check if we can found in the root gopath
		_, err := os.Stat(src)
		if err != nil {
			continue
		}

		dst := filepath.Join(b.TmpDir, "src", dep)

		if err := copy.Copy(src, dst, copy.Options{Skip: skipCopy}); err != nil {
			logger.Errorf("Failed to Copy the folder from %v to %v, the error is: %v ", src, dst, err)
		}

		visited[src] = true
	}
}
func skipCopy(src string, info os.FileInfo) (bool, error) {
	irregularModeType := os.ModeNamedPipe | os.ModeSocket | os.ModeDevice | os.ModeCharDevice | os.ModeIrregular
	if strings.HasSuffix(src, "/.git") {
		logger.Infof("Skip .git dir [%s]", src)
		return true, nil
	}
	if info.Mode()&irregularModeType != 0 {
		logger.Warnf("Skip file [%s], the file mode is [%s]", src, info.Mode().String())
		return true, nil
	}
	return false, nil
}
func (b *Build) cpNonStandardLegacy() {
	for _, v := range b.Pkgs {
		if v.Name == "main" {
			dst := b.TmpDir
			src := v.Dir

			if err := copy.Copy(src, dst, copy.Options{Skip: skipCopy}); err != nil {
				logger.Errorf("Failed to Copy the folder from %v to %v, the error is: %v ", src, dst, err)
			}
			break
		}
	}
}
