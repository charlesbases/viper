package website

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"testing"
)

func TestDuration(t *testing.T) {
	{
		var d Duration = 1
		fmt.Println(d.Encode())
	}

	{
		var d Duration = 60
		fmt.Println(d.Encode())
	}

	{
		var d Duration = 61
		fmt.Println(d.Encode())
	}

	{
		var d Duration = 3600
		fmt.Println(d.Encode())
	}
}

func Test(t *testing.T) {
	err := filepath.WalkDir("../resources", func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			fmt.Println(path)
			return fs.SkipAll
		}
		return nil
	})
	if err != nil {
		fmt.Println(err)
	}
}
