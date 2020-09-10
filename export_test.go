package modver

import (
	"sync"
	"testing"
)

var allVersionMutex sync.Mutex

func SetAllVersion(t *testing.T, vers []ModuleVersion) {
	t.Helper()
	allVersionMutex.Lock()
	t.Cleanup(func() {
		allVersion = AllVersion
		allVersionMutex.Unlock()
	})

	allVersion = func(module string) ([]ModuleVersion, error) {
		return vers, nil
	}
}
