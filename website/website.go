package website

import (
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/pkg/errors"

	"github.com/charlesbases/viper/logger"
)

// ErrLinkType 错误的视频链接
var ErrLinkType = func(link Link) error { return errors.Errorf("invalid link of %s", link) }

// rootDir 视频资源文件夹
var rootDir = func() string {
	abs, err := filepath.Abs("resources")
	logger.Error(errors.Wrap(err, "abs of resources"))
	return abs
}()

// RsResolution .
type RsResolution struct {
	// rule 分辨率排序规则
	rule func(v string) int
	// list 分辨率列表
	list []string
}

// Add .
func (r *RsResolution) Add(v string) {
	r.list = append(r.list, v)
}

// Best .
func (r *RsResolution) Best() string {
	if len(r.list) != 0 {
		// 根据 rule 规则对 list 排序
		sort.Slice(r.list, func(i, j int) bool {
			return r.rule(r.list[i]) > r.rule(r.list[j])
		})
		return r.list[0]
	}
	return ""
}

// NewResolutionRule .
func NewResolutionRule(rule func(v string) int) *RsResolution {
	return &RsResolution{rule: rule, list: make([]string, 0, 8)}
}

// RsInfo .
type RsInfo struct {
	RootDir  string
	Uploader string

	// err error of os.MkdirAll
	err  error
	once sync.Once
}

// folder .
func (inf *RsInfo) folder() string {
	return filepath.Join(rootDir, inf.RootDir, inf.Uploader)
}

// mkdir .
func (inf *RsInfo) mkdir() error {
	inf.once.Do(func() {
		inf.err = os.MkdirAll(inf.folder(), 0755)
		logger.Error(inf.err)
	})
	return inf.err
}

// RsVideo .
type RsVideo struct {
	Inf *RsInfo

	// ID 视频 ID
	ID string
	// Hlink hls link
	Hlink Link
	// Parts parts of video
	Parts []Link
}

// absname .
func (v *RsVideo) absname() string {
	return filepath.Join(v.Inf.folder(), v.ID+".mkv")
}

// download .
func (v *RsVideo) download() error {
	if err := v.Inf.mkdir(); err != nil {
		return err
	}

	// 视频下载
	file, err := os.OpenFile(v.absname(), os.O_CREATE, 0644)
	if err != nil {
		return err
	}

	for _, part := range v.Parts {
		if err := part.Fetch(WriteTo(file)); err != nil {
			file.Close()
			os.Remove(v.absname())
			return err
		}
	}

	return file.Close()
}

// IsNotExist .
func (v *RsVideo) IsNotExist() bool {
	_, err := os.Stat(v.absname())
	return err != nil
}

// WebHook .
type WebHook interface {
	LinkType() LinkType
	Resources() (chan *RsVideo, error)
}

// H .
func H(hook WebHook) error {
	if videos, err := hook.Resources(); err != nil {
		return err
	} else {
		for video := range videos {
			logger.Error(video.download())
		}
	}
	return nil
}
