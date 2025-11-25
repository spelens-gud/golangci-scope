package build

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/rogpeppe/go-internal/modfile"
	"github.com/spelens-gud/logger"
	"github.com/tongjingran/copy"
)

func (b *Build) cpGoModulesProject() {
	for _, v := range b.Pkgs {
		if v.Name == "main" {
			dst := b.TmpDir
			src := v.Module.Dir

			if err := copy.Copy(src, dst, copy.Options{Skip: skipCopy}); err != nil {
				logger.Errorf("Failed to Copy the folder from %v to %v, the error is: %v ", src, dst, err)
			}
			break
		} else {
			continue
		}
	}
}
func (b *Build) updateGoModFile() (updateFlag bool, newModFile []byte, err error) {
	// use buildflags `-mod=vendor` and exist vendor folder, should not update go.mod
	if _, err1 := os.Stat(path.Join(b.ModRoot, "vendor")); err1 == nil && strings.Contains(b.BuildFlags, "-mod=vendor") {
		return
	}
	tempModfile := filepath.Join(b.TmpDir, "go.mod")
	buf, err := ioutil.ReadFile(tempModfile)
	if err != nil {
		return
	}
	oriGoModFile, err := modfile.Parse(tempModfile, buf, nil)
	if err != nil {
		return
	}

	updateFlag = false
	for index := range oriGoModFile.Replace {
		replace := oriGoModFile.Replace[index]
		oldPath := replace.Old.Path
		oldVersion := replace.Old.Version
		newPath := replace.New.Path
		newVersion := replace.New.Version
		// replace to a local filesystem does not have a version
		// absolute path no need to rewrite
		if newVersion == "" && !filepath.IsAbs(newPath) {
			var absPath string
			fullPath := filepath.Join(b.ModRoot, newPath)
			absPath, _ = filepath.Abs(fullPath)
			// DropReplace & AddReplace will not return error
			// so no need to check the error
			_ = oriGoModFile.DropReplace(oldPath, oldVersion)
			_ = oriGoModFile.AddReplace(oldPath, oldVersion, absPath, newVersion)
			updateFlag = true
		}
	}
	oriGoModFile.Cleanup()
	// Format will not return error, so ignore the returned error
	// func (f *File) Format() ([]byte, error) {
	//     return Format(f.Syntax), nil
	// }
	newModFile, _ = oriGoModFile.Format()
	return
}
