package utils

import (
	"fmt"
	connector "github.com/ayhanozemre/fs-shadow/path"
	"os"
	"path/filepath"
	"testing"
)

func Test_Sum(t *testing.T) {
	testFolder := "/tmp/fs-shadow"
	_ = os.Mkdir(testFolder, os.ModePerm)

	emptyFile := filepath.Join(testFolder, "test.txt")
	_, _ = os.Create(emptyFile)

	emptyFolder := filepath.Join(testFolder, "test")
	_ = os.Mkdir(emptyFolder, os.ModePerm)

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
		{Name: "test-1", Path: emptyFile, Expected: "4233d3a7e7fd4f4002438e06cda20cebcd406d8fdae2903b191bb1405fece2d2"},
		{Name: "test-2", Path: emptyFolder, Expected: "ab755dd2ab5e07da5c2f5059c577b6a8f4344ead4eab65035bd70b2bec8fff39"},
		{Name: "test-3", Path: fullFolder, Expected: "a6ed0c785d4590bc95c216bcf514384eee6765b1c2b732d0b0a1ad7e14d3204a"},
		{Name: "test-4", Path: fullFile, Expected: "f5a50145968c188060b73b4f8d7042ce6a880df2186825b6716b50342010ff6c"},
	}
	for _, c := range cases {
		sum, err := Sum(connector.NewFSPath(c.Path))
		if err != nil {
			fmt.Println("err")
		}
		if sum != c.Expected {
			t.Fatalf("[%s] Results:%s and ExpectResult:%s not equal", c.Name, sum, c.Expected)
		}
	}

	_ = os.Remove(testFolder)

}
