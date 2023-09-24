package website

import (
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
