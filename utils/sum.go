package utils

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/ayhanozemre/fs-shadow/path"
	"io"
	"io/ioutil"
	"os"
)

func Sum(path connector.Path) (string, error) {
	if path.IsDir() {
		return FolderSum(path.String())
	}
	return FileSum(path.String())
}

func FolderSum(path string) (string, error) {
	deepLimit := 100
	deepCount := 0
	var buff bytes.Buffer

	files, _ := ioutil.ReadDir(path)
	for _, p := range files {
		if deepLimit == deepCount {
			break
		}
		buff.WriteString(p.Name())
		deepCount += 1
	}
	if buff.String() == "" {
		// If the folder is empty, use created at.
		p, _ := os.Stat(path)
		buff.WriteString(fmt.Sprint(p.ModTime().Unix()))
	}
	hasher := sha256.New()
	hasher.Write(buff.Bytes())
	value := hex.EncodeToString(hasher.Sum(nil))
	return value, nil
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
