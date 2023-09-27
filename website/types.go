package website

import (
	"strconv"
	"strings"
)

const (
	// LinkTypeUnknown 未知的链接
	LinkTypeUnknown LinkType = iota
	// LinkTypeVideo 视频链接
	LinkTypeVideo
	// LinkTypeUploader 创作者链接
	LinkTypeUploader
)

// LinkType 链接类型
type LinkType int8

// Link 链接
type Link string

// Join .
func (l Link) Join(v ...string) Link {
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

// Fetch .
func (l Link) Fetch(fn reader, opts ...func(meta *Metadata)) error {
	return fetch(l, fn, opts...)
}

// String .
func (l Link) String() string {
	return string(l)
}

const (
	// Second 秒
	Second Duration = 1
	// Minute 分
	Minute = 60 * Second
	// Hour 时
	Hour = 60 * Minute
)

// Duration 视频时长
type Duration int64

// Encode .
// 返回 1'30'30 形式的时长字符串
// 30     表示 30s
// 2'30   表示 2m30s
// 1‘2’30 表示 1h2m30s
func (d Duration) Encode() string {
	if d == 0 {
		return ""
	}

	// s
	if d < Minute {
		return strconv.Itoa(int(d))
	}
	// m
	if d < Hour {
		var m, s = d / Minute, d % Minute
		return strings.Join([]string{strconv.Itoa(int(m)), strconv.Itoa(int(s))}, "'")
	}
	// 	h
	var h, m, s = d / Hour, d % Hour / Minute, d % Minute
	return strings.Join([]string{strconv.Itoa(int(h)), strconv.Itoa(int(m)), strconv.Itoa(int(s))}, "'")
}
