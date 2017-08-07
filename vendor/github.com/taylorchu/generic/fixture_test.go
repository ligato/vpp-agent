package generic

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func testRewritePackage(t *testing.T, pkgPath, newPkgPath string, typeMap map[string]Target, expect string) {
	testRewritePackageWithInput(t, pkgPath, newPkgPath, typeMap, "", expect)
}

func testRewritePackageWithInput(t *testing.T, pkgPath, newPkgPath string, typeMap map[string]Target, input, expect string) {
	const dirname = "rewrite_test"
	err := os.MkdirAll(dirname, 0777)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dirname)

	if input != "" {
		err = copyDir(dirname, input)
		if err != nil {
			t.Fatal(err)
		}
	}

	if strings.HasPrefix(newPkgPath, ".") {
		os.Setenv("GOPACKAGE", "GOPACKAGE")
		defer os.Unsetenv("GOPACKAGE")
	}

	err = os.Chdir(dirname)
	if err != nil {
		t.Fatal(err)
	}

	ctx, err := NewContext(pkgPath, newPkgPath)
	if err != nil {
		t.Fatal(err)
	}
	ctx.TypeMap = typeMap

	err = RewritePackage(ctx)
	os.Chdir("..")
	if err != nil {
		t.Fatal(err)
	}

	assertEqualDir(t, dirname, expect)
}

func copyDir(to, from string) error {
	fi, err := ioutil.ReadDir(from)
	if err != nil {
		return err
	}
	for _, info := range fi {
		if info.IsDir() {
			continue
		}

		tof, err := os.Create(filepath.Join(to, info.Name()))
		if err != nil {
			return err
		}
		defer tof.Close()

		fromf, err := os.Open(filepath.Join(from, info.Name()))
		if err != nil {
			return err
		}
		defer fromf.Close()

		_, err = io.Copy(tof, fromf)
		if err != nil {
			return err
		}
	}
	return nil
}

func assertEqualDir(t *testing.T, path1, path2 string) {
	t.Log(path1, path2)
	fi1, err := ioutil.ReadDir(path1)
	if err != nil {
		t.Fatal(err)
	}
	fi2, err := ioutil.ReadDir(path2)
	if err != nil {
		t.Fatal(err)
	}
	if len(fi1) != len(fi2) {
		t.Fatalf("%s: %d, %s: %d", path1, len(fi1), path2, len(fi2))
	}

	for _, info := range fi1 {
		p1 := filepath.Join(path1, info.Name())
		p2 := filepath.Join(path2, info.Name())
		if info.IsDir() {
			assertEqualDir(t, p1, p2)
		} else {
			b1, err := ioutil.ReadFile(p1)
			if err != nil {
				t.Fatal(err)
			}
			b2, err := ioutil.ReadFile(p2)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(b1, b2) {
				t.Fatalf("\n%s:\n\n%s\n%s:\n\n%s\n", p1, b1, p2, b2)
			}
		}
	}
}
