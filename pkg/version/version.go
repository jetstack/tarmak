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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime"
	"strings"

	"github.com/blang/semver"
	apimachineryversion "k8s.io/apimachinery/pkg/version"
)

const (
	releasesURL = "https://api.github.com/repos/jetstack/tarmak/releases"
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
		for _, preRelease := range []string{"alpha", "beta", "rc"} {

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

type Tag struct {
	TagName string `json:"tag_name"`
}

func LastVersion() (string, error) {
	resp, err := http.Get(releasesURL)
	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	tags := new([]Tag)
	err = json.Unmarshal(body, tags)
	if err != nil {
		return "", fmt.Errorf("error unmarshalling body '%s': %s", body, err)
	}

	var versions semver.Versions
	for _, tag := range *tags {
		v, err := semver.Make(tag.TagName)
		if err != nil {
			return "", fmt.Errorf("failed to parse release tag: %s", err)
		}

		versions = append(versions, v)
	}

	return lastVersionFromVersions(versions)
}

func lastVersionFromVersions(versions semver.Versions) (string, error) {
	if versions.Len() < 1 {
		return "", fmt.Errorf(
			"no versions found in tarmak releases %s", releasesURL)
	}

	semver.Sort(versions)
	currVersion, err := semver.Make(CleanVersion())
	if err != nil {
		return "", fmt.Errorf("failed to parse current version %s: %s", CleanVersion(), err)
	}

	lastVersion, err := semver.Make("0.0.0")
	if err != nil {
		return "", err
	}

	for _, v := range versions {
		if v.GT(lastVersion) && len(v.Pre) == 0 &&
			v.Minor < currVersion.Minor {
			lastVersion = v
		}
	}

	return lastVersion.String(), nil
}
