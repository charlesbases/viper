package rp

import (
	"fmt"
	"regexp"
)

var (
	// TypeVideo 视频链接
	TypeVideo = regexp.MustCompile(`https://www.xvideos.com/video[0-9]+/_`)
	// TypeUploader 艺术家链接
	TypeUploader = regexp.MustCompile(`https://www.xvideos.com/amateur-channels/[a-zA-Z0-9_-]+`)
)

var (
	// VideoSuffix 根据艺术家主页的视频列表，获取视频链接
	VideoSuffix = func(uploader string) *regexp.Regexp { return regexp.MustCompile(fmt.Sprintf(`.*/%s/(.+)`, uploader)) }
	// VideoUploader 根据视频链接，获取艺术家
	VideoUploader = regexp.MustCompile(`html5player[.]setUploaderName[(]'(.*)'[)];`)
	// VideoDuration 获取视频时长
	VideoDuration = regexp.MustCompile(`<meta property="og:duration" content="(.*)" />`)
	// VideoHls 获取视频下载链接
	VideoHls = regexp.MustCompile(`html5player[.]setVideoHLS[(]'(.*)/hls.m3u8'[)];`)
	// VideoID 视频 ID
	VideoID = regexp.MustCompile(`https://www[.]xvideos[.]com/video([0-9]+)`)
)
