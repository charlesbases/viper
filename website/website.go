package website

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charlesbases/progressbar"
	"github.com/charlesbases/salmon"
	"github.com/pkg/errors"

	"github.com/charlesbases/viper/logger"
)

// format 视频文件格式
const format = "mkv"

// root 视频资源文件夹
var root = func() string {
	if abs, err := filepath.Abs("resource"); err != nil {
		panic(err)
	} else {
		return abs
	}
}()

var resource = []string{}

func init() {
	filepath.Walk(
		root, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return filepath.SkipAll
			}

			if !info.IsDir() {
				resource = append(resource, filepath.Base(path))
			}
			return nil
		},
	)
}

// 并发下载数
var concurrent = 10

// SetConcurrent .
func SetConcurrent(c int) {
	if c > 0 {
		concurrent = c
	}
}

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
		sort.Slice(
			r.list, func(i, j int) bool {
				return r.rule(r.list[i]) > r.rule(r.list[j])
			},
		)
		return r.list[0]
	}
	return ""
}

// NewResolutionRule .
func NewResolutionRule(rule func(v string) int) *RsResolution {
	return &RsResolution{rule: rule, list: make([]string, 0, 8)}
}

// Link 链接
type Link string

// String .
func (l Link) String() string {
	return string(l)
}

// Fetch .
func (l Link) Fetch(fn reader, opts ...func(meta *Metadata)) error {
	return fetch(l, fn, opts...)
}

// Joins .
func (l Link) Joins(v ...string) Link {
	var br strings.Builder
	var n = len(l.String())

	for _, val := range v {
		n += len(val)
	}

	br.Grow(n)
	br.WriteString(l.String())

	for _, val := range v {
		br.WriteString("/")
		br.WriteString(strings.TrimPrefix(val, "/"))
	}
	return Link(br.String())
}

// RsVideo .
type RsVideo struct {
	RsVideoHeader

	// 视频文件路径
	abspath string
	// HLink hls link
	HLink Link
	// Parts of video
	Parts []Link
}

// RsVideoHeader .
type RsVideoHeader struct {
	// 视频 ID
	ID string
	// 网页链接
	WebLink Link
}

// RsInfor .
type RsInfor struct {
	// 所有者
	Owner string
	// 视频网站首页
	WebHome Link
	// 下载文件夹
	rootDir string
	// 视频列表
	VideoList []*RsVideoHeader
}

// mkdirall .
func (infor *RsInfor) mkdirall() error {
	infor.rootDir = filepath.Join(root, infor.WebHome.String(), infor.Owner)
	return os.MkdirAll(infor.rootDir, 0755)
}

// exists .
func (infor *RsInfor) exists(header *RsVideoHeader) bool {
	if len(header.ID) == 0 {
		logger.Errorf("%s: id is empty", header.WebLink)
		return true
	}

	for _, path := range resource {
		if strings.HasPrefix(path, header.ID) {
			return true
		}
	}

	return false
}

// ready .
func (infor *RsInfor) ready(video *RsVideo) (*os.File, error) {
	if len(infor.Owner) == 0 {
		return nil, errors.New("unknown uploader")
	}
	if len(video.Parts) == 0 {
		return nil, errors.New("unknown parts")
	}

	video.abspath = filepath.Join(infor.rootDir, video.ID)

	return os.OpenFile(video.abspath, os.O_CREATE, 0644)
}

// rollback .
func (infor *RsInfor) rollback(video *RsVideo) {
	os.Remove(video.abspath)
}

// commit .
func (infor *RsInfor) commit(video *RsVideo) {
	target := strings.Join([]string{video.abspath, format}, ".")
	os.Rename(video.abspath, target)

	resource = append(resource, target)
}

// WebHook .
type WebHook interface {
	// LinkList 视频列表
	LinkList() (*RsInfor, error)
	// VideoLink 根据视频网页链接，获取下载地址
	VideoLink(header *RsVideoHeader) (*RsVideo, error)
}

// H .
func H(reader *progressbar.Reader, hook WebHook) error {
	infor, err := hook.LinkList()
	if err != nil {
		return err
	}

	pb := reader.NewProgress(infor.WebHome.Joins(infor.Owner).String(), uint(len(infor.VideoList)))

	// 创建文件夹
	if err := infor.mkdirall(); err != nil {
		return err
	}

	pool, err := salmon.NewPool(
		concurrent, func(v interface{}, stop func()) {
			if header, ok := v.(*RsVideoHeader); ok {
				if infor.exists(header) {
					goto finshed
				}

				// 根据视频网页链接，解析下载地址
				if video, err := hook.VideoLink(header); err != nil {
					logger.Error(errors.Wrap(err, header.WebLink.String()))
				} else {
					// 创建视频文件
					if file, err := infor.ready(video); err != nil {
						logger.Error(errors.Wrap(err, header.WebLink.String()))
					} else {
						// 视频分片下载
						var derr error
						for _, part := range video.Parts {
							if derr = part.Fetch(WriteTo(file)); derr != nil {
								logger.Error(errors.Wrap(err, header.WebLink.String()))
								break
							}
						}
						file.Close()

						if derr != nil {
							infor.rollback(video)
						} else {
							infor.commit(video)
						}
					}
				}

			finshed:
				pb.Incr(1)
				// <-time.After(time.Second)
			}
		},
	)
	if err != nil {
		return err
	}

	// 视频下载
	for _, header := range infor.VideoList {
		pool.Invoke(header)
	}

	pool.Wait()
	return nil
}
