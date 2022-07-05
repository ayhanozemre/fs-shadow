package utils

import (
	connector "github.com/ayhanozemre/fs-shadow/path"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func Test_Sum(t *testing.T) {
	testFolder := "/tmp/fs-shadow"
	_ = os.Mkdir(testFolder, os.ModePerm)

	emptyFile := filepath.Join(testFolder, "test.txt")
	_, _ = os.Create(emptyFile)

	fullFolder := filepath.Join(testFolder, "test-full")
	_ = os.Mkdir(fullFolder, os.ModePerm)
	subfile := filepath.Join(fullFolder, "test.txt")
	_, _ = os.Create(subfile)

	fullFile := filepath.Join(testFolder, "test-full.txt")
	file, _ := os.Create(fullFile)
	_, _ = file.WriteString("file is not empty")
	_ = file.Close()

	type TestCase struct {
		Name     string
		Path     string
		Expected string
	}
	cases := []TestCase{
		{Name: "test-1", Path: emptyFile, Expected: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
		{Name: "test-2", Path: fullFolder, Expected: "a6ed0c785d4590bc95c216bcf514384eee6765b1c2b732d0b0a1ad7e14d3204a"},
		{Name: "test-3", Path: fullFile, Expected: "f5a50145968c188060b73b4f8d7042ce6a880df2186825b6716b50342010ff6c"},
	}
	for _, c := range cases {
		sum, err := Sum(connector.NewFSPath(c.Path))
		assert.Equal(t, nil, err, "sum error")

		if sum != c.Expected {
			t.Fatalf("[%s] Results:%s and ExpectResult:%s not equal", c.Name, sum, c.Expected)
		}
	}

	_ = os.RemoveAll(testFolder)

}
