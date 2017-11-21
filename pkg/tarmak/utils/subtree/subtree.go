// Copyright Jetstack Ltd. See LICENSE for details.

// This implement utilities around git subtree
package subtree

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Sirupsen/logrus"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

// This methods isolates a sub folder of the git and return its hash
func SubtreeSplit(rootDir string, prefix string) (commitHash string, err error) {
	command := []string{"git", "subtree", "split", "--prefix", prefix}
	cmd := exec.Command(command[0], command[1:]...)
	cmd.Dir = rootDir

	var stdOut bytes.Buffer
	var stdErr bytes.Buffer
	cmd.Stdout = &stdOut
	cmd.Stderr = &stdErr

	err = cmd.Run()
	if err != nil {
		return "", fmt.Errorf(
			"cmd=%+v rootDir=%s stderr=%s",
			command,
			rootDir,
			strings.TrimSpace(stdErr.String()),
		)
	}

	return strings.TrimSpace(stdOut.String()), nil
}

type Subtree struct {
	log     *logrus.Entry
	rootDir *string

	Prefix           string
	RemoteRepository string
	RemoteRef        string
}

func New(prefix, remoteRepository string) *Subtree {
	log := logrus.New()
	if testing.Verbose() {
		log.Level = logrus.DebugLevel
	} else {
		log.Out = ioutil.Discard
	}
	return &Subtree{
		log:              log.WithField("prefix", prefix),
		Prefix:           prefix,
		RemoteRepository: remoteRepository,
		RemoteRef:        "master",
	}
}

// this return the root dir of the git repository
func (s *Subtree) RootDir() (string, error) {
	if s.rootDir != nil {
		return *s.rootDir, nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("couldn't get current working directory: %s", err)
	}

	path, err := filepath.Abs(filepath.Join(cwd, ".."))
	if err != nil {
		return "", fmt.Errorf("couldn't determine absoulte path: %s", err)
	}
	s.rootDir = &path

	return path, nil
}

// build a nice and easy remote name
func (s *Subtree) RemoteName() string {
	return strings.Replace(s.Prefix, "/", "-", -1)
}

// this methods tests if subtree vendored repository has been upsteamed. This makes sure that all puppet modules changes have to go to their repository first before they can be merge into tarmak
func (s *Subtree) TestSubtreeUpstream(t *testing.T) {
	gitPath, err := s.RootDir()
	if err != nil {
		t.Fatal(err)
	}

	subtreeHash, err := SubtreeSplit(gitPath, s.Prefix)
	if err != nil {
		t.Fatalf("unable to determine subtree's hash: %s", err)
	}
	s.log.WithField("hash", subtreeHash).Info("subtree hash")

	g, err := git.PlainOpen(gitPath)
	if err != nil {
		t.Fatalf("git open: %s", err)
	}

	remotes, err := g.Remotes()
	if err != nil {
		t.Fatalf("error reading remotes: %s", err)
	}

	// find or create remote
	var remote *git.Remote
	remoteName := s.RemoteName()
	for pos, _ := range remotes {
		if remotes[pos].Config().Name == remoteName {
			remote = remotes[pos]
			break
		}
	}

	var remoteFetched bool
	remoteFetchOptions := &git.FetchOptions{Tags: git.NoTags}

	if remote == nil {
		remote, err = g.CreateRemote(&config.RemoteConfig{
			Name: remoteName,
			URLs: []string{s.RemoteRepository},
			Fetch: []config.RefSpec{
				config.RefSpec(fmt.Sprintf("+refs/heads/*:refs/remotes/%s/*", remoteName)),
			},
		})
		if err != nil {
			t.Fatalf("error creating remote: %s", err)
		}

		if err := remote.Fetch(remoteFetchOptions); err != nil {
			t.Fatalf("error fetching remote: %s", err)
		}
		remoteFetched = true
	}

	for {

		refName := plumbing.ReferenceName(filepath.Join("refs/remotes", remoteName, s.RemoteRef))
		ref, err := g.Reference(refName, true)
		if err != nil {
			t.Fatalf("error resolving ref '%s': %s", refName, err)
		}

		found, err := commitInRemote(g, ref, subtreeHash)
		if err != nil {
			t.Fatalf("error iterating ref '%s': %s", refName, err)
		}

		if found {
			return
		}

		// if haven't fetched, fetch now and retest
		if !remoteFetched {
			if err := remote.Fetch(remoteFetchOptions); err != nil {
				break
			}
			remoteFetched = true
			continue
		}

		break

	}

	t.Fatalf(
		"subtree not found upstream commit=%s prefix=%s remote=%s remote_ref=%s",
		subtreeHash,
		s.Prefix,
		s.RemoteRepository,
		s.RemoteRef,
	)
}

// walk through commit in remote to find reference
func commitInRemote(g *git.Repository, ref *plumbing.Reference, commitHash string) (found bool, err error) {
	commitIter, err := g.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return false, err
	}

	for {
		commit, err := commitIter.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return false, err
		}
		if commit.Hash.String() == commitHash {
			return true, nil
		}
	}

	return false, nil
}
