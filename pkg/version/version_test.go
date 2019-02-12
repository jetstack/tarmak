// Copyright Jetstack Ltd. See LICENSE for details.
package version

import (
	"testing"
)

func Test_CleanVersion(t *testing.T) {
	for _, test := range []struct {
		version, minor, exp string
	}{
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
	} {

		gitVersion = test.version
		gitMinor = test.minor

		if got := CleanVersion(); got != test.exp {
			t.Errorf("clear version failed, exp=%s\ngot=%s", test, got)
		}
	}
}
