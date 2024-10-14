package main

import (
	iofs "io/fs"
	"slices"
	"testing"
)

// todo: test for ???:
//   files, err := getAllFilenames(staticFS)
//   log.Printf("  debug fsys:\n    %v\n    %s\n", files, err)

func Test_ExpectedEmbededFiles(t *testing.T) {
	expectedFiles := []string{
		"web/static/generated/main.js",
		"web/static/generated/output.css",
	}

	embededFiles, err := getAllFilenames(fs)
	if err != nil {
		t.Errorf("getAllFilenames() error = %v", err)
	}

	for _, expectedFile := range expectedFiles {
		if !slices.Contains(embededFiles, expectedFile) {
			t.Errorf("expected embeded files, got '%v', want '%v' but was not found", embededFiles, expectedFile)
		}
	}

}

func getAllFilenames(efs iofs.FS) (files []string, err error) {
	if err := iofs.WalkDir(efs, ".", func(path string, d iofs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		files = append(files, path)

		return nil
	}); err != nil {
		return nil, err
	}

	return files, nil
}
