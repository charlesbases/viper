package xvideos

import (
	"path/filepath"
	"strconv"
	"strings"

	"github.com/charlesbases/viper/logger"
	"github.com/charlesbases/viper/website"
	"github.com/charlesbases/viper/website/xvideos/rp"
)

const (
	// rootDir 下载文件夹
	rootDir = "xvideos.com"
	// rootHome 视频网站主页
	rootHome website.Link = "https://www.xvideos.com"
)

var bestResolution = func(v string) int {
	prefixs := []string{"hls-1080p", "hls-720p"}
	for idx, prefix := range prefixs {
		if strings.HasPrefix(v, prefix) {
			return len(prefixs) - idx
		}
	}
	return -1
}

type hook struct {
	link website.Link
}

// WebHome 网站首页
func (h *hook) WebHome() string {
	return rootDir
}

// LinkType 链接类型
func (h *hook) LinkType() website.LinkType {
	if rp.TypeVideo.MatchString(h.link.String()) {
		return website.LinkTypeVideo
	}
	if rp.TypeUploader.MatchString(h.link.String()) {
		return website.LinkTypeUploader
	}
	return website.LinkTypeUnknown
}

// Resources 资源列表
func (h *hook) Resources() (*website.RsInfo, error) {
	var inf = website.NewRsInfo(rootDir)

	switch h.LinkType() {
	case website.LinkTypeVideo:
		inf.Total(1)

		go func() {
			if video := h.video(inf, h.link); video != nil {
				inf.Push(video)
			}
			inf.Close()
		}()

		return inf, nil
	case website.LinkTypeUploader:
		return h.videos(inf)
	default:
		return nil, website.ErrLinkType(h.link)
	}
}

// UploaderResponse 艺术家主页
type UploaderResponse struct {
	Videos []*struct {
		U string `json:"u"`
	} `json:"videos"`
}

// videos 艺术家主页视频解析
func (h *hook) videos(inf *website.RsInfo) (*website.RsInfo, error) {
	h.link = website.Link(strings.Split(h.link.String(), "#")[0])

	inf.Uploader = filepath.Base(h.link.String())

	var suffixs = make([]string, 0)

	{
		var page int
		for {
			var resp = new(UploaderResponse)
			if err := h.link.Join("videos", "new", strconv.Itoa(page)).Fetch(website.Unmarshal(resp)); err != nil {
				logger.Error(err)
			} else {
				if len(resp.Videos) == 0 {
					break
				}
				for _, video := range resp.Videos {
					if suffix := website.FindSubstring(rp.VideoSuffix(inf.Uploader), video.U); len(suffix) != 0 {
						suffixs = append(suffixs, "video"+suffix)
					}
				}
				page++
			}
		}
	}

	inf.Total(len(suffixs))

	go func() {
		for _, suffix := range suffixs {
			if video := h.video(inf, rootHome.Join(suffix)); video != nil {
				inf.Push(video)
			}
		}

		inf.Close()
	}()

	return inf, nil
}

// video 视频链接解析
func (h *hook) video(inf *website.RsInfo, link website.Link) *website.RsVideo {
	if id := website.FindSubstring(rp.VideoID, link.String()); len(id) != 0 && !inf.IsExist(id) {
		var video = &website.RsVideo{ID: id, Link: link}

		// hls
		if err := link.Fetch(website.ReadLine(func(line string) (isBreak bool) {
			if video.Duration == 0 {
				video.Duration = website.Duration(website.FindSubnumber(rp.VideoDuration, line))
			}
			if len(video.Hlink) == 0 {
				video.Hlink = website.Link(website.FindSubstring(rp.VideoHls, line))
			}
			if len(inf.Uploader) == 0 {
				inf.Uploader = website.FindSubstring(rp.VideoUploader, line)
			}
			return len(inf.Uploader) != 0 && len(video.Hlink) != 0
		})); err != nil {
			logger.Error(err)
		}

		// parts
		if len(video.Hlink) != 0 {
			h.parts(video)
			return video
		}
	}
	return nil
}

// parts 视频分片链接解析
func (h *hook) parts(video *website.RsVideo) {
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
}

// H .
func H(link website.Link) website.WebHook {
	return &hook{link: link}
}
