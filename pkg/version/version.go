// Copyright Jetstack Ltd. See LICENSE for details.
/*
Copyright 2014 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package version

import (
	"fmt"
	"runtime"
	"strings"

	apimachineryversion "k8s.io/apimachinery/pkg/version"
)

// Get returns the overall codebase version. It's for detecting
// what code a binary was built from.
func Get() apimachineryversion.Info {
	// These variables typically come from -ldflags settings and in
	// their absence fallback to the settings in pkg/version/base.go
	return apimachineryversion.Info{
		Major:        gitMajor,
		Minor:        gitMinor,
		GitVersion:   gitVersion,
		GitCommit:    gitCommit,
		GitTreeState: gitTreeState,
		BuildDate:    buildDate,
		GoVersion:    runtime.Version(),
		Compiler:     runtime.Compiler,
		Platform:     fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

func CleanVersion() string {
	if !strings.HasSuffix(gitMinor, "+") {
		return gitVersion
	}

	var out []string
	split := strings.Split(gitVersion, "-")

LOOP:
	for _, s := range split {
		for _, preRelease := range []string{"alpha", "beta"} {

			if strings.Contains(s, preRelease) {
				out = append(out, strings.Split(s, ".")[0])
				break LOOP
			}
		}

		if strings.HasPrefix(s, "dirty") {
			break
		}

		out = append(out, s)
	}

	return strings.Join(out, "-")
}
