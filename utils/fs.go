package utils

import (
	"embed"
	"fmt"
	"os"
	"path"
)

func copyDir(fs embed.FS, src, dest string) error {
	entries, err := fs.ReadDir(src)
	if err != nil {
		return fmt.Errorf("failed to read directory '%s' from embed.FS", src)
	}

	for _, e := range entries {
		if e.IsDir() {
			err = os.MkdirAll(path.Join(dest, e.Name()), os.ModePerm)
			if err != nil {
				return fmt.Errorf("failed to create directory")
			}
			copyDir(fs, path.Join(src, e.Name()), path.Join(dest, e.Name()))
		} else {
			fpath := path.Join(src, e.Name())
			in, err := fs.ReadFile(fpath)
			if err != nil {
				return fmt.Errorf("failed to read file %s", fpath)
			}

			outPath := path.Join(dest, e.Name())
			err = os.WriteFile(outPath, in, os.ModePerm)
			if err != nil {
				return fmt.Errorf("failed to write file %s", outPath)
			}
		}
	}

	return nil
}

func CopyDirFromEmbedFS(fs embed.FS, src, dest string) error {
	return copyDir(fs, src, dest)
}
