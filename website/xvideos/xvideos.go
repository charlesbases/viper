package xvideos

import (
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/charlesbases/viper/logger"
	"github.com/charlesbases/viper/website"
	"github.com/charlesbases/viper/website/xvideos/rp"
)

var bestResolution = func(v string) int {
	prefixs := []string{"hls-1080p"}
	for idx, prefix := range prefixs {
		if strings.HasPrefix(v, prefix) {
			return len(prefixs) - idx
		}
	}
	return -1
}

// xvideos .
type xvideos struct {
	res website.Link
	inf *website.RsInfo
}

// WebHome 网站首页
func (x *xvideos) WebHome() website.Link {
	return "https://www.xvideos.com"
}

// UploaderResponse 艺术家主页
type UploaderResponse struct {
	Videos []*struct {
		U string `json:"u"`
	} `json:"videos"`
}

// LinkList 视频链接列表
func (x *xvideos) LinkList() *website.RsInfo {
	x.inf.LinkList = make([]*website.RsVideoDesc, 0)
	// 视频链接
	if rp.TypeVideo.MatchString(x.res.String()) {
		x.inf.LinkList = append(x.inf.LinkList,
			&website.RsVideoDesc{
				ID:   website.FindSubstring(rp.VideoID, x.res.String()),
				Link: x.res,
			})
		return x.inf
	}

	// 艺术家链接
	if rp.TypeUploader.MatchString(x.res.String()) {
		x.inf.Uploader = website.FindSubstring(rp.VideoUploaderWithLink, x.res.String())

		var page int
		for {
			var resp = new(UploaderResponse)
			if err := x.res.Join("videos", "new", strconv.Itoa(page)).Fetch(website.Unmarshal(resp)); err != nil {
				logger.Error(err)
				break
			} else {
				if len(resp.Videos) == 0 {
					break
				}
				for _, video := range resp.Videos {
					if suffix := website.FindSubstring(rp.VideoSuffix, video.U); len(suffix) != 0 {
						link := x.WebHome().Join("video" + suffix)
						x.inf.LinkList = append(x.inf.LinkList,
							&website.RsVideoDesc{
								ID:   website.FindSubstring(rp.VideoID, link.String()),
								Link: link,
							})
					}
				}
				page++
			}
		}

		return x.inf
	}

	logger.Errorf("%s: unknown link type", x.res)
	return nil
}

// ParseVideo 根据视频链接，获取下载地址
func (x *xvideos) ParseVideo(v *website.RsVideoDesc) (*website.RsVideo, error) {
	if len(v.ID) != 0 {
		var video = &website.RsVideo{RsVideoDesc: *v}

		// hls
		if err := v.Link.Fetch(website.ReadLine(func(line string) (isBreak bool) {
			if video.Duration == 0 {
				video.Duration = website.Duration(website.FindSubnumber(rp.VideoDuration, line))
			}
			if len(video.Hlink) == 0 {
				video.Hlink = website.Link(website.FindSubstring(rp.VideoHls, line))
			}
			if len(x.inf.Uploader) == 0 {
				x.inf.Uploader = website.FindSubstring(rp.VideoUploaderWithHTML, line)
			}
			return len(x.inf.Uploader) != 0 && len(video.Hlink) != 0
		})); err != nil {
			logger.Error(err)
		}

		// duration
		if video.Duration < 120 {
			return nil, errors.New("often less than 120 seconds")
		}

		// parts
		if len(video.Hlink) != 0 {
			return x.videoParts(video), nil
		}
	}
	return nil, errors.New("unknown id for this video")
}

// videoParts .
func (x *xvideos) videoParts(video *website.RsVideo) *website.RsVideo {
	var rst = website.NewResolutionRule(bestResolution)

	// 分辨率
	video.Hlink.Join("hls.m3u8").Fetch(website.ReadLine(func(line string) (isBreak bool) {
		if !strings.HasPrefix(line, "#") {
			rst.Add(line)
		}
		return false
	}))

	// 	视频下载列表
	video.Parts = make([]website.Link, 0)
	video.Hlink.Join(rst.Best()).Fetch(website.ReadLine(func(line string) (isBreak bool) {
		if !strings.HasPrefix(line, "#") {
			video.Parts = append(video.Parts, video.Hlink.Join(line))
		}
		return false
	}))

	return video
}

// New .
func New(res website.Link) website.WebHook {
	return &xvideos{res: res, inf: &website.RsInfo{RootDir: "xvideos.com"}}
}
