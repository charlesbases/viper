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

func TestVideoSuffix(t *testing.T) {
	var str = `/prof-video-click/model/chicken1806/78063627/_`
	fmt.Println(VideoSuffix.FindStringSubmatch(str))
}
