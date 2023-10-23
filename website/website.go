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

// RsInfo .
type RsInfo struct {
	// RootDir 视频下载文件夹
	RootDir string
	// Uploader 艺术家
	Uploader string
	// LinkList 视频列表
	LinkList []*RsVideoDesc
}

// bar 打印进度条
func (inf *RsInfo) bar(count, total int) {
	fmt.Printf("\r%s/%s: [%d/%d]", inf.RootDir, inf.Uploader, count, total)
}

// mkdir .
func (inf *RsInfo) mkdir() error {
	return os.MkdirAll(filepath.Join(root, inf.RootDir, inf.Uploader), 0755)
}

// isExist .
func (inf *RsInfo) isExist(v *RsVideoDesc) (isExist bool) {
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

// abspath 视频文件全路径
func (inf *RsInfo) abspath(v *RsVideo) (string, error) {
	if len(inf.Uploader) == 0 {
		return "", errors.New("unknown uploader")
	}
	if v.Duration == 0 {
		return "", errors.New("unknown video duration")
	}
	if len(v.Parts) == 0 {
		return "", errors.New("unknown video parts")
	}
	return filepath.Join(root, inf.RootDir, inf.Uploader, strings.Join([]string{v.ID, v.Duration.Encode()}, "_")), nil
}

// RsVideoDesc .
type RsVideoDesc struct {
	// ID 视频 ID
	ID string
	// Link 网页视频链接
	Link Link
}

// RsVideo .
type RsVideo struct {
	RsVideoDesc
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
	// WebHome 视频网站首页
	WebHome() Link
	// LinkList 视频链接列表
	LinkList() *RsInfo
	// ParseVideo 根据视频链接，获取下载地址
	ParseVideo(v *RsVideoDesc) (*RsVideo, error)
}

// H .
func H(hook WebHook) error {
	var works = make(chan struct{}, concurrent)
	for i := 0; i < concurrent; i++ {
		works <- struct{}{}
	}

	var lock sync.Mutex

	var info = hook.LinkList()
	var total, count = len(info.LinkList), 0

	if err := info.mkdir(); err != nil {
		return err
	}

	var wgroup sync.WaitGroup
	wgroup.Add(total)

	info.bar(count, total)
	for _, val := range info.LinkList {
		v := val
		go func() {
			<-works

			if info.isExist(v) {
				goto finshed
			}

			// 根据视频网页链接，解析下载地址
			if video, err := hook.ParseVideo(v); err != nil {
				logger.Error(errors.Wrap(err, v.Link.String()))
			} else {
				// 获取视频下载文件名
				if abspath, err := info.abspath(video); err != nil {
					logger.Error(errors.Wrap(err, v.Link.String()))
				} else {
					// 创建视频文件
					if file, err := os.OpenFile(abspath, os.O_CREATE, 0644); err != nil {
						logger.Error(errors.Wrap(err, v.Link.String()))
					} else {
						var derr error

						// 视频下载
						for _, part := range video.Parts {
							if derr = part.Fetch(WriteTo(file)); derr != nil {
								logger.Error(errors.Wrap(err, v.Link.String()))
								break
							}
						}

						file.Close()
						if derr != nil {
							os.Remove(abspath)
						} else {
							os.Rename(abspath, strings.Join([]string{abspath, format}, "."))
						}
					}
				}
			}

		finshed:
			lock.Lock()
			count++
			info.bar(count, total)
			lock.Unlock()

			wgroup.Done()
			works <- struct{}{}

		}()
	}

	wgroup.Wait()
	fmt.Println(" [ok]")
	return nil
}
