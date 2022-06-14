package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/ayhanozemre/fs-shadow/connector"
	"io"
	"os"
)

func Sum(path connector.Path) (string, error) {
	if path.IsDir() {
		// folder sum is not necessary for now, but it should be here as an idea.
		return FolderSum(path.String())
	}
	return FileSum(path.String())
}

func FolderSum(path string) (string, error) {
	// sum of sums of sub folders can be hashed
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
