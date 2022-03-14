package testing

import (
	"os"
	"testing"
)

func AssertNotErr(t *testing.T, err error) {
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func AssertDirExists(t *testing.T, dir string) {
	f, err := os.Stat(dir)
	if os.IsNotExist(err) {
		t.Errorf("%s does not exist", dir)
	}
	if !f.IsDir() {
		t.Errorf("%s is not a dir", dir)
	}
}

func AssertDirNotExists(t *testing.T, dir string) {
	_, err := os.Stat(dir)
	if err != nil && !os.IsNotExist(err) {
		t.Errorf("dir %s exists, it should not", dir)
	}
}

func AssertFileExists(t *testing.T, file string) {
	f, err := os.Stat(file)
	if err != nil {
		t.Errorf("%s does not exist", file)
	}
	if f.IsDir() {
		t.Errorf("%s is not a file", file)
	}
}

func AssertFileNotExists(t *testing.T, file string) {
	if _, err := os.Stat(file); !os.IsNotExist(err) {
		t.Errorf("%s does exists, it should not", file)
	}
}
