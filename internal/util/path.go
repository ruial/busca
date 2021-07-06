package util

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"
)

// SafeJoin joins two paths and the resulting path always has the base prefix to prevent directory traversal
// Alternative implementation: https://github.com/opencontainers/runc/blob/master/libcontainer/utils/utils.go
func SafeJoin(baseDir, file string) (string, error) {
	out := path.Join(baseDir, file)
	prefixAbs, err := filepath.Abs(baseDir)
	if err != nil {
		return out, err
	}
	outAbs, err := filepath.Abs(out)
	if err != nil {
		return out, err
	}
	if !strings.HasPrefix(outAbs, prefixAbs) {
		return out, fmt.Errorf("File path %s is causing directory traversal", file)
	}
	return out, nil
}
