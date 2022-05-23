package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
)

func Sum(path string) (string, error) {
	p := Path(path)
	if p.IsDir() {
		// folder sum is not necessary for now, but it should be here as an idea.
		return FolderSum(p.String())
	}
	return FileSum(p.String())
}

func FolderSum(path string) (string, error) {
	return "", nil
}

func FileSum(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, f); err != nil {
		return "", err
	}
	value := hex.EncodeToString(hasher.Sum(nil))
	return value, nil
}
