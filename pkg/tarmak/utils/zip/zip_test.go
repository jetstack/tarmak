// Copyright Jetstack Ltd. See LICENSE for details.
package zip

import (
	"archive/zip"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"testing"
)

func Test_Zip(t *testing.T) {
	dir, err := ioutil.TempDir("", "tarmak_zip_test")
	try(t, nil, err)
	defer os.RemoveAll(dir)

	// can't override source
	src := []string{"a.zip"}
	dst := "a"
	try(t, fmt.Errorf(overrideErr, src, src[0]), Zip(src, dst, false))
	dst = "a.zip"
	try(t, fmt.Errorf(overrideErr, src, dst), Zip(src, dst, false))

	tmpSrc := tempFile(t, dir)
	b := randBytes(t)
	try(t, nil, ioutil.WriteFile(tmpSrc, b, 0664))
	src = []string{tmpSrc}
	dst = fmt.Sprintf("%s.zip", tempFile(t, dir))

	// shouldn't delete source
	try(t, nil, Zip(src, dst, false))
	_, err = os.Stat(tmpSrc)
	try(t, nil, err)

	// should delete source
	try(t, nil, Zip(src, dst, true))
	_, err = os.Stat(tmpSrc)
	try(t, fmt.Errorf("stat %s: no such file or directory", tmpSrc), err)

	testUnzip(t, b, dst)

	// should still unzip correctly when destination target is source file
	tmpSrc = tempFile(t, dir)
	tmpDst := tempFile(t, dir)
	b = randBytes(t)
	try(t, nil, ioutil.WriteFile(tmpSrc, b, 0664))
	try(t, nil, ioutil.WriteFile(tmpDst, b, 0664))
	try(t, nil, Zip([]string{tmpDst}, tmpDst, true))

	tmpDst = fmt.Sprintf("%s.zip", tmpDst)
	testUnzip(t, b, tmpDst)

	src = []string{tmpSrc, tmpDst}

	// zip source into itself
	try(t, nil, Zip(src, tmpDst, true))
	_, err = os.Stat(tmpDst)
	try(t, nil, err)

	testUnzip(t, b, tmpDst)
}

func testUnzip(t *testing.T, srcB []byte, dst string) {
	reader, err := zip.OpenReader(dst)
	try(t, nil, err)

	if len(reader.File) == 0 {
		t.Fatalf("expected zip to contain at least 1 file, got=%d", len(reader.File))
	}

	f, err := reader.File[0].Open()
	try(t, nil, err)
	b, err := ioutil.ReadAll(f)
	try(t, nil, err)

	if string(b) != string(srcB) {
		t.Fatalf("final bytes don't match, exp=%s got=%s", srcB, b)
	}
}

func try(t *testing.T, exp, got error) {
	err := fmt.Sprintf("unexpected result, exp=[%v] got=[%v]",
		exp, got)

	if exp == got {
		return
	}

	if exp == nil && got != nil {
		t.Fatal(err)
	}

	if exp != nil && got == nil {
		t.Fatal(err)
	}

	if exp.Error() != got.Error() {
		t.Fatal(err)
	}
}

func tempFile(t *testing.T, dir string) string {
	f, err := ioutil.TempFile(dir, "")
	if err != nil {
		t.Fatalf("failed to create temp file: %s", err)
	}

	return f.Name()
}

func randBytes(t *testing.T) []byte {
	var b [256]byte
	_, err := rand.Read(b[:])
	if err != nil {
		t.Fatalf("failed to generate random bytes: %s", err)
	}

	return b[:]
}
