package common

import (
	"fmt"
	"io/ioutil"
	"os"
	"ox/pkg/util/oregexp"
	"path/filepath"
	"strings"
)

// GetModPath ...
func GetModPath(projectPath string) (modPath string) {
	dir := filepath.Dir(projectPath)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			content, _ := ioutil.ReadFile(filepath.Join(dir, "go.mod"))
			mod := oregexp.RegexpReplace(`module\s+(?P<name>[\S]+)`, string(content), "$name")
			name := strings.TrimPrefix(filepath.Dir(projectPath), dir)
			name = strings.TrimPrefix(name, string(os.PathSeparator))
			if name == "" {
				return fmt.Sprintf("%s/", mod)
			}
			return fmt.Sprintf("%s/%s/", mod, name)
		}
		parent := filepath.Dir(dir)
		if dir == parent {
			return ""
		}
		dir = parent
	}
}
