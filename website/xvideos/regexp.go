package xvideos

import (
	"fmt"
	"regexp"
)

var (
	// linkTypeVideo 视频链接
	linkTypeVideo = regexp.MustCompile(`https://www.xvideos.com/video[0-9]+/_`)
	// linkTypeUploader 艺术家链接
	linkTypeUploader = regexp.MustCompile(`https://www.xvideos.com/amateur-channels/[a-zA-Z0-9_-]+`)
)

var (
	// regexpVideoSuffix 根据艺术家主页的视频列表，获取视频链接
	regexpVideoSuffix = func(uploader string) *regexp.Regexp { return regexp.MustCompile(fmt.Sprintf(`.*/%s/(.+)`, uploader)) }
	// regexpVideoUploader 根据视频链接，获取艺术家
	regexpVideoUploader = regexp.MustCompile(`html5player[.]setUploaderName[(]'(.*)'[)];`)
	// regexpVideoHls 获取视频下载链接
	regexpVideoHls = regexp.MustCompile(`html5player[.]setVideoHLS[(]'(.*)/hls.m3u8'[)];`)
	// regexpVideoID 视频 ID
	regexpVideoID = regexp.MustCompile(`https://www[.]xvideos[.]com/video([0-9]+)`)
)

// findSubstring .
func findSubstring(r *regexp.Regexp, s string) string {
	if subs := r.FindStringSubmatch(s); len(subs) > 1 {
		return subs[1]
	}
	return ""
}
