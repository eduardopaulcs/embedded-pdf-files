package service

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/fs"
)

func ComputeStaticHash(fsys fs.FS) string {
	h := sha256.New()
	fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		f, err := fsys.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		io.Copy(h, f)
		return nil
	})
	return fmt.Sprintf("%x", h.Sum(nil)[:8])
}
