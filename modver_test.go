package modver_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/tenntenn/modver"
)

func modvers(t *testing.T, vers []string) []modver.ModuleVersion {
	t.Helper()
	if vers == nil {
		return nil
	}
	modvers := make([]modver.ModuleVersion, len(vers))
	for i := range modvers {
		modvers[i] = modver.ModuleVersion{
			Module:  "example.com/sample",
			Version: vers[i],
		}
	}
	return modvers
}

func TestFilterVersion(t *testing.T) {
	t.Parallel()

	type S = []string
	cases := []struct {
		versions    []string
		constraints string
		want        []string
	}{
		{S{"v1.0.0", "v1.1.0", "v1.2.0"}, ">= v1.1.0", S{"v1.1.0", "v1.2.0"}},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.constraints, func(t *testing.T) {
			t.Parallel()
			modver.SetAllVersion(t, modvers(t, tt.versions))
			got, err := modver.FilterVersion("example.com/sample", tt.constraints)
			if err != nil {
				t.Error("unexpected error:", err)
			}

			if diff := cmp.Diff(modvers(t, tt.want), got); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestLatestVersion(t *testing.T) {
	t.Parallel()

	type S = []string
	cases := map[string]struct {
		versions    []string
		max int
		want        []string
	}{
		"even": {S{"v1.0.0", "v1.0.1", "v1.1.0", "v1.1.1"}, 2, S{"v1.0.1", "v1.1.1"}},
		"odd": {S{"v1.0.0", "v1.0.1", "v1.1.0", "v1.1.1", "v1.2.2"}, 3, S{"v1.0.1", "v1.1.1", "v1.2.2"}},
		"notenough": {S{"v1.0.0", "v1.0.1"}, 2, S{"v1.0.1"}},
		"over": {S{"v1.0.0", "v1.0.1", "v1.1.0", "v1.1.1"}, 1, S{"v1.1.1"}},
		"empty": {nil, 1, nil},
		"zero": {S{"v1.0.0", "v1.0.1", "v1.1.0", "v1.1.1"}, 0, nil},
	}

	for name, tt := range cases {
		tt := tt
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			modver.SetAllVersion(t, modvers(t, tt.versions))
			got, err := modver.LatestVersion("example.com/sample", tt.max)
			if err != nil {
				t.Error("unexpected error:", err)
			}

			if diff := cmp.Diff(modvers(t, tt.want), got); diff != "" {
				t.Error(diff)
			}
		})
	}
}
