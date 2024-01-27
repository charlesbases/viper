package xvideos

import (
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/charlesbases/viper/logger"
	"github.com/charlesbases/viper/website"
)

var _ website.WebHook = (*xvideos)(nil)

const home website.Link = "https://www.xvideos.com"

var resolution = func(v string) int {
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
	src   website.Link
	infor *website.RsInfor
}

// VideoListResponse 艺术家主页
type VideoListResponse struct {
	Videos []*struct {
		U string `json:"u"`
	} `json:"videos"`
}

// LinkList 视频列表
func (x *xvideos) LinkList() (*website.RsInfor, error) {
	switch {
	// video
	case regexpIsVideo.MatchString(x.src.String()):
		x.infor.VideoList = append(
			x.infor.VideoList, &website.RsVideoHeader{
				ID:      website.FindSubstring(regexpVideoID, x.src.String()),
				WebLink: x.src,
			},
		)
		return x.infor, nil
	// owner
	case regexpIsOwner.MatchString(x.src.String()):
		x.infor.Owner = website.FindSubstring(regexpVideoOwner, x.src.String())

		// 获取视频列表
		var page int
		for {
			var resp = new(VideoListResponse)
			// 格式错误
			if err := x.src.Joins("videos", "new", strconv.Itoa(page)).Fetch(website.Unmarshal(resp)); err != nil {
				return nil, err
			}
			// 页码错误
			if len(resp.Videos) == 0 {
				break
			}

			for _, video := range resp.Videos {
				if suffix := website.FindSubstring(regexpVideoSuffix, video.U); len(suffix) != 0 {
					weblink := home.Joins("video" + suffix)

					x.infor.VideoList = append(
						x.infor.VideoList,
						&website.RsVideoHeader{
							ID:      website.FindSubstring(regexpVideoID, weblink.String()),
							WebLink: weblink,
						},
					)
				}
			}
			page++
		}
		return x.infor, nil
	default:
		return nil, errors.Errorf("%s: unknown link type", x.src)
	}
}

// VideoLink 根据视频网页链接，获取下载地址
func (x *xvideos) VideoLink(header *website.RsVideoHeader) (*website.RsVideo, error) {
	video := &website.RsVideo{RsVideoHeader: *header}
	// hls link
	if err := header.WebLink.Fetch(
		website.ReadLine(
			func(line string) (isBreak bool) {
				if len(video.HLink) == 0 {
					video.HLink = website.Link(website.FindSubstring(regexpVideoHLink, line))
				}
				if len(x.infor.Owner) == 0 {
					x.infor.Owner = website.FindSubstring(regexpVideoOwnerWithHTML, line)
				}
				return len(x.infor.Owner) != 0 && len(video.HLink) != 0
			},
		),
	); err != nil {
		logger.Error(errors.Wrap(err, header.WebLink.String()))
	}

	if len(video.HLink) != 0 {
		return x.videoParts(video), nil
	}

	return nil, errors.Errorf("%s: not found", header.WebLink)
}

// videoParts .
func (x *xvideos) videoParts(video *website.RsVideo) *website.RsVideo {
	var rst = website.NewResolutionRule(resolution)

	// 分辨率
	video.HLink.Joins("hls.m3u8").Fetch(
		website.ReadLine(
			func(line string) (isBreak bool) {
				if !strings.HasPrefix(line, "#") {
					rst.Add(line)
				}
				return false
			},
		),
	)

	// 	视频下载列表
	video.Parts = make([]website.Link, 0)
	video.HLink.Joins(rst.Best()).Fetch(
		website.ReadLine(
			func(line string) (isBreak bool) {
				if !strings.HasPrefix(line, "#") {
					video.Parts = append(video.Parts, video.HLink.Joins(line))
				}
				return false
			},
		),
	)

	return video
}

// New .
func New(src website.Link) website.WebHook {
	return &xvideos{
		src: src,
		infor: &website.RsInfor{
			WebHome: "xvideos.com",
		},
	}
}
