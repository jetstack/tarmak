// Copyright Jetstack Ltd. See LICENSE for details.
package version

import (
	"testing"

	"github.com/blang/semver"
)

func Test_CleanVersion(t *testing.T) {
	for _, test := range []struct {
		version, minor, exp string
	}{
		{
			version: "0.6.0-rc1.15+337112cd82fadb]",
			minor:   "6+",
			exp:     "0.6.0-rc1",
		},

		{
			version: "0.6.0-alpha2.6+b364c17a97386f-dirty",
			minor:   "6+",
			exp:     "0.6.0-alpha2",
		},

		{
			version: "0.6.0-alpha2-dirty",
			minor:   "6+",
			exp:     "0.6.0-alpha2",
		},

		{
			version: "0.6.0-alpha2",
			minor:   "6+",
			exp:     "0.6.0-alpha2",
		},

		{
			version: "0.6.0",
			minor:   "6",
			exp:     "0.6.0",
		},

		{
			version: "0.6.1",
			minor:   "6",
			exp:     "0.6.1",
		},

		{
			version: "0.6.0-dirty",
			minor:   "6+",
			exp:     "0.6.0",
		},

		{
			version: "0.6.0-rc1",
			minor:   "6",
			exp:     "0.6.0-rc1",
		},
	} {

		gitVersion = test.version
		gitMinor = test.minor

		if got := CleanVersion(); got != test.exp {
			t.Errorf("clear version failed, exp=%s got=%s", test.exp, got)
		}
	}
}

func Test_LastVersion(t *testing.T) {
	_, err := LastVersion()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := lastVersionFromVersions(semver.Versions{}); err == nil {
		t.Fatalf("expected error, got=%s", err)
	}

	gitVersion = "0.6.1"
	gitMinor = "6"

	versionsStr := []string{
		"0.6.1",
		"0.6.0",
		"0.6.0-alpha1",
		"0.6.0-rc1",
		"0.4.1",
		"0.4.0",
		"0.4.0-rc1",
		"0.2.1",
		"0.2.1-rc2",
		"0.1.1",
	}

	var versions semver.Versions
	for _, vStr := range versionsStr {
		v, err := semver.Make(vStr)
		if err != nil {
			t.Fatal(err)
		}

		versions = append(versions, v)
	}

	lastVersion, err := lastVersionFromVersions(versions)
	if err != nil {
		t.Fatalf("unexpected error, got=%s", err)
	}

	if lastVersion != "0.4.1" {
		t.Fatalf("expected 0.4.1, got=%s, %s", lastVersion, CleanVersion())
	}

	gitVersion = "0.4.1"
	gitMinor = "4"

	lastVersion, err = lastVersionFromVersions(versions)
	if err != nil {
		t.Fatalf("unexpected error, got=%s", err)
	}

	if lastVersion != "0.2.1" {
		t.Fatalf("expected 0.2.1, got=%s", lastVersion)
	}
}
