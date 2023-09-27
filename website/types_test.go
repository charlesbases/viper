package website

import (
	"fmt"
	"strings"
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
	fmt.Println(strings.Join([]string{"12345", Duration(0).Encode(), format}, "."))
}
