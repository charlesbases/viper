package rp

import (
	"fmt"
	"regexp"
	"testing"
)

func TestVideoDuration(t *testing.T) {
	var str = `<meta property="og:duration" content="542" />`
	fmt.Println(regexp.MustCompile(`<meta property="og:duration" content="(.*)" />`).FindStringSubmatch(str))
}
