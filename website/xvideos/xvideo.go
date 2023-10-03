package xvideos

import (
	"strconv"

	"github.com/charlesbases/viper/logger"
	"github.com/charlesbases/viper/website"
	"github.com/charlesbases/viper/website/xvideos/rp"
)

// xvideos .
type xvideos struct {
	res website.Link
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
func (x *xvideos) LinkList() []website.Link {
	// 视频链接
	if rp.TypeVideo.MatchString(x.res.String()) {
		return []website.Link{x.res}
	}

	// 艺术家链接
	if rp.TypeUploader.MatchString(x.res.String()) {
		var links = make([]website.Link, 0)

		var page int
		for {
			var resp = new(UploaderResponse)
			if err := x.res.Join("videos", "new", strconv.Itoa(page)).Fetch(website.Unmarshal(resp)); err != nil {
				logger.Error(err)
			} else {
				if len(resp.Videos) == 0 {
					break
				}
				for _, video := range resp.Videos {
					if suffix := website.FindSubstring(rp.VideoSuffix, video.U); len(suffix) != 0 {
						links = append(links, x.WebHome().Join("video"+suffix))
					}
				}
				page++
			}
		}

		return links
	}

	logger.Errorf("%s: unknown link type")
	return nil
}

// ParseVideo 根据视频链接，获取下载地址
func (x *xvideos) ParseVideo(link website.Link) (*website.RsVideo, error) {
	// TODO implement me
	panic("implement me")
}

// Resources 视频资源
func (x *xvideos) Resources() (*website.rsInfo, error) {
	// TODO implement me
	panic("implement me")
}

// New .
func New(res website.Link) website.WebHook {
	return &xvideos{res: res}
}
