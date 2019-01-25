// Copyright Jetstack Ltd. See LICENSE for details.
package zip

import (
	"bytes"
	"os"
	"reflect"
	"testing"
)

func TestZipBytes(t *testing.T) {

	doZip := func() ([]byte, error) {
		buf := new(bytes.Buffer)
		err := ZipBytes(
			[]string{"test-1/file-1.txt", "test-2/file-2.txt", "test-1/test-2/file12.txt"},
			[][]byte{[]byte("1\n"), []byte("2\n"), []byte("12-secret\n")},
			[]os.FileMode{0644, 0644, 0600},
			buf,
		)
		if err != nil {
			return []byte{}, err
		}
		return buf.Bytes(), nil
	}

	zip1, err1 := doZip()
	if err1 != nil {
		t.Errorf("unexpected error: %v", err1)
	}

	zip2, err2 := doZip()
	if err2 != nil {
		t.Errorf("unexpected error: %v", err2)
	}

	if !reflect.DeepEqual(zip1, zip2) {
		t.Error("both zip files of the same files did mismatch")
	}
}
