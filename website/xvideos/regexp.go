package xvideos

import "regexp"

var (
	// 视频链接
	regexpIsVideo = regexp.MustCompile(`https://www.xvideos.com/video[0-9]+/_`)
	// 艺术家链接
	regexpIsOwner = regexp.MustCompile(`https://www.xvideos.com/.*/[a-zA-Z0-9_-]+`)
)

var (
	// 根据视频链接获取视频 ID
	regexpVideoID = regexp.MustCompile(`https://www[.]xvideos[.]com/video([0-9]+)`)
)

var (
	// 根据视频链接，获取艺术家
	regexpVideoOwner = regexp.MustCompile(`https://www.xvideos.com/.*/(.*)`)
	// 解析首页的视频列表，获取视频 ID
	regexpVideoSuffix = regexp.MustCompile(`.*/([1-9]+.*)`)
	// 解析首页的视频列表，获取视频下载链接
	regexpVideoHLink = regexp.MustCompile(`html5player[.]setVideoHLS[(]'(.*)/hls.m3u8'[)];`)
	// 解析首页的视频列表，获取艺术家
	regexpVideoOwnerWithHTML = regexp.MustCompile(`html5player[.]setUploaderName[(]'(.*)'[)];`)
)
