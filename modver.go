package modver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/hashicorp/go-version"
)

// It can create a mock.
var allVersion = AllVersion

// ModuleVersion has module path and its version.
type ModuleVersion struct {
	Module  string
	Version string
}

// String implements fmt.Stringer.
func (modver ModuleVersion) String() string {
	return fmt.Sprintf("%s@%s", modver.Module, modver.Version)
}

// AllVersion get available all versions of the module.
func AllVersion(module string) ([]ModuleVersion, error) {
	dir, err := ioutil.TempDir("", "modver*")
	if err != nil {
		return nil, fmt.Errorf("cannot create tmpdir: %w", err)
	}
	defer os.RemoveAll(dir)

	if _, err := execCmd(dir, "go", "mod", "init", "tmp"); err != nil {
		return nil, fmt.Errorf("go mod init tmp: %w", err)
	}

	r, err := execCmd(dir, "go", "list", "-m", "-versions", "-json", module)
	if err != nil {
		return nil, fmt.Errorf("go list -m -versions -json: %w", err)
	}

	var v struct{ Versions []string }
	if err := json.NewDecoder(r).Decode(&v); err != nil {
		return nil, fmt.Errorf("cannot decode JSON: %w", err)
	}

	vers := make([]ModuleVersion, len(v.Versions))
	for i := range v.Versions {
		vers[i] = ModuleVersion{
			Module:  module,
			Version: v.Versions[i],
		}
	}

	return vers, nil
}

// FilterVersion returns versions of the module which satisfy the constraints such as ">= v2.0.0"
// The constraints rule uses github.com/hashicorp/go-version.
func FilterVersion(module, constraints string) ([]ModuleVersion, error) {
	c, err := version.NewConstraint(constraints)
	if err != nil {
		return nil, fmt.Errorf("cannot parse constraints: %w", err)
	}

	all, err := allVersion(module)
	if err != nil {
		return nil, err
	}

	var vers []ModuleVersion
	for _, ver := range all {
		v, err := version.NewVersion(ver.Version)
		if err != nil {
			return nil, fmt.Errorf("cannot parse version: %w", err)
		}
		if c.Check(v) {
			vers = append(vers, ver)
		}
	}

	return vers, nil
}

// LatestVersion returns most latest versions (<= max) of each minner version.
func LatestVersion(module string, max int) ([]ModuleVersion, error) {
	if max <= 0 {
		return nil, nil
	}

	all, err := allVersion(module)
	if err != nil {
		return nil, err
	}

	var vers []ModuleVersion
	minors := map[int]bool{}
	for i := len(all) - 1; i >= 0; i-- {
		v, err := version.NewVersion(all[i].Version)
		if err != nil {
			return nil, fmt.Errorf("cannot parse version: %w", err)
		}

		segs := v.Segments()
		if minors[segs[1]] {
			continue
		}
		minors[segs[1]] = true

		vers = append(vers, all[i])
		if len(vers) >= max {
			break
		}
	}

	for i := 0; i < len(vers)/2; i++ {
		vers[i], vers[len(vers)-(i+1)] = vers[len(vers)-(i+1)], vers[i]
	}

	return vers, nil
}

func execCmd(dir, cmd string, args ...string) (io.Reader, error) {
	var stdout, stderr bytes.Buffer
	_cmd := exec.Command(cmd, args...)
	_cmd.Stdout = &stdout
	_cmd.Stderr = &stderr
	_cmd.Dir = dir
	if err := _cmd.Run(); err != nil {
		return nil, fmt.Errorf("%w:\n%s", err, &stderr)
	}
	return &stdout, nil
}
