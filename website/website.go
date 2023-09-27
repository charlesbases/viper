package website

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/pkg/errors"

	"github.com/charlesbases/viper/logger"
)

// format 视频文件格式
const format = "mkv"

// concurrent 并发下载数
var concurrent = 10

var (
	// ErrLinkType 错误的视频链接
	ErrLinkType = func(link Link) error { return errors.Errorf("invalid link of %s", link) }
)

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
	// RootDir 视频文件夹
	RootDir string
	// Uploader 创作者
	Uploader string
	// videoc videos chan
	videoc chan *RsVideo

	// total 视频数量
	total int
	// count 已下载视频数
	count int

	lock  sync.RWMutex
	group sync.WaitGroup
}

// NewRsInfo .
func NewRsInfo(root string) *RsInfo {
	return &RsInfo{RootDir: root, videoc: make(chan *RsVideo, concurrent)}
}

// Total .
func (inf *RsInfo) Total(total int) {
	inf.total = total
}

// Push .
func (inf *RsInfo) Push(video *RsVideo) {
	inf.videoc <- video
}

// Close .
func (inf *RsInfo) Close() {
	close(inf.videoc)
}

// IsExist 视频文件是否下载
func (inf *RsInfo) IsExist(id string) bool {
	if len(inf.Uploader) != 0 {
		if entry, err := os.ReadDir(inf.folder()); err != nil {
			logger.Error(err)
			return false
		} else {
			for _, val := range entry {
				if !val.IsDir() && strings.HasPrefix(val.Name(), id) {
					return true
				}
			}
			return false
		}
	}
	logger.Errorf("%s: unknown video uploader")
	return true
}

// videopath return video path
func (inf *RsInfo) videopath(v *RsVideo) string {
	if v.Duration != 0 {
		return filepath.Join(inf.folder(), strings.Join([]string{v.ID, v.Duration.Encode(), format}, "."))
	}

	logger.Errorf("%s: unknown video duration")
	return filepath.Join(inf.folder(), strings.Join([]string{v.ID, format}, "."))
}

// folder .
func (inf *RsInfo) folder() string {
	return filepath.Join(rootDir, inf.RootDir, inf.Uploader)
}

// done .
func (inf *RsInfo) done() {
	inf.group.Add(-1)

	inf.lock.Lock()
	inf.count++
	inf.lock.Unlock()

	inf.bar()
}

// bar .
func (inf *RsInfo) bar() {
	inf.lock.RLock()
	fmt.Printf("\r%s/%s: [%d/%d]", inf.RootDir, inf.Uploader, inf.count, inf.total)
	inf.lock.RUnlock()
}

// download .
func (inf *RsInfo) download() error {
	if err := os.MkdirAll(inf.folder(), 0755); err != nil {
		return err
	}

	inf.group.Add(inf.total)

	var works = make(chan struct{}, concurrent)
	for i := 0; i < concurrent; i++ {
		works <- struct{}{}
	}

	for video := range inf.videoc {
		inf.bar()

		go func(v *RsVideo) {
			<-works
			var name = inf.videopath(v)
			if f, err := os.OpenFile(name, os.O_CREATE, 0644); err != nil {
				logger.Error(err)
			} else {
				for _, part := range v.Parts {
					if err := part.Fetch(WriteTo(f)); err != nil {
						f.Close()
						os.Remove(name)
						logger.Error(err)
						goto finish
					}
				}
				f.Close()
			}

		finish:
			works <- struct{}{}
			inf.done()
		}(video)
	}

	inf.group.Wait()
	close(works)

	fmt.Println(" [ok]")
	return nil
}

// RsVideo .
type RsVideo struct {
	// ID 视频 ID
	ID string
	// Link 网页视频链接
	Link Link
	// Hlink hls link
	Hlink Link
	// Duration 视频时长
	Duration Duration
	// Parts parts of video
	Parts []Link
}

// SetConcurrent .
func SetConcurrent(c int) {
	if c != 0 {
		concurrent = c
	}
}

// WebHook .
type WebHook interface {
	WebHome() string
	LinkType() LinkType
	Resources() (*RsInfo, error)
}

// H .
func H(hook WebHook) error {
	if inf, err := hook.Resources(); err != nil {
		return err
	} else {
		return inf.download()
	}
}
