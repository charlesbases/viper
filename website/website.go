package website

import (
	"fmt"
	"io/fs"
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

// root 视频资源文件夹
var root = func() string {
	if abs, err := filepath.Abs("resources"); err != nil {
		panic(err)
	} else {
		return abs
	}
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

// rsInfo .
type rsInfo struct {
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
func NewRsInfo(root string) *rsInfo {
	return &rsInfo{RootDir: root, videoc: make(chan *RsVideo, concurrent)}
}

// Total .
func (inf *rsInfo) Total(total int) {
	inf.total = total
}

// Push .
func (inf *rsInfo) Push(video *RsVideo) {
	inf.videoc <- video
}

// Close .
func (inf *rsInfo) Close() {
	close(inf.videoc)
}

// videopath return video path
func (inf *rsInfo) videopath(v *RsVideo) string {
	if v.Duration != 0 {
		return filepath.Join(inf.folder(), strings.Join([]string{v.ID, v.Duration.Encode(), format}, "."))
	}

	logger.Errorf("%s: unknown video duration")
	return filepath.Join(inf.folder(), strings.Join([]string{v.ID, format}, "."))
}

// folder .
func (inf *rsInfo) folder() string {
	return filepath.Join(root, inf.RootDir, inf.Uploader)
}

// done .
func (inf *rsInfo) done() {
	inf.group.Add(-1)

	inf.lock.Lock()
	inf.count++
	inf.lock.Unlock()

	inf.bar()
}

// bar .
func (inf *rsInfo) bar() {
	inf.lock.RLock()
	fmt.Printf("\r%s/%s: [%d/%d]", inf.RootDir, inf.Uploader, inf.count, inf.total)
	inf.lock.RUnlock()
}

// download .
func (inf *rsInfo) download() error {
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

// IsExist .
func (v *RsVideo) IsExist() (isExist bool) {
	if len(v.ID) == 0 {
		logger.Errorf("%s: id is empty", v.Link)
		return true
	}

	filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() && strings.HasPrefix(filepath.Base(path), v.ID) {
			isExist = true
			return fs.SkipAll
		}
		return nil
	})

	return
}

// SetConcurrent .
func SetConcurrent(c int) {
	if c != 0 {
		concurrent = c
	}
}

// WebHook .
type WebHook interface {
	// WebHome 视频网站首页
	WebHome() Link
	// LinkList 视频链接列表
	LinkList() []Link

	// ParseVideo 根据视频链接，获取下载地址
	ParseVideo(link Link) (*RsVideo, error)
	// Resources 视频资源
	Resources() (*rsInfo, error)
}

// H .
func H(hook WebHook) error {
	var works = make(chan struct{}, concurrent)
	for i := 0; i < concurrent; i++ {
		works <- struct{}{}
	}

	for _, val := range hook.LinkList() {
		link := val
		go func() {
			<-works
			video, err := hook.ParseVideo(link)
			if err != nil {
				logger.Error()
			}
		}()
	}
}
