package website

import (
	"bufio"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
)

var ErrNotFound = errors.New("not found")

var defaultClient = &http.Client{Timeout: 30 * time.Second}

var defaultHeader = map[string]string{
	// "Content-Type": "application/json",
	// "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36",
	"User-Agent": "Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko; compatible; Googlebot/2.1; +http://www.google.com/bot.html) Chrome/W.X.Y.Z Safari/537.36",
}

type reader func(r io.Reader) error

type Metadata struct {
	Method string
	Header map[string]string
	Body   io.Reader
}

// fetch .
func fetch(l Link, fn reader, opts ...func(meta *Metadata)) error {
	var meta = &Metadata{Method: http.MethodGet, Header: make(map[string]string, 0)}
	for _, opt := range opts {
		opt(meta)
	}

	req, err := http.NewRequest(meta.Method, l.String(), meta.Body)
	if err != nil {
		return errors.Wrap(err, "http")
	}

	for key, val := range defaultHeader {
		req.Header.Add(key, val)
	}
	for key, val := range meta.Header {
		req.Header.Add(key, val)
	}

	resp, err := defaultClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "http")
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return ErrNotFound
	}

	return fn(resp.Body)
}

// ReadLine .
func ReadLine(fn func(line string) (isBreak bool)) reader {
	return func(r io.Reader) error {
		var buf = bufio.NewReader(r)
		for {
			if line, err := buf.ReadString('\n'); err != nil {
				break
			} else {
				if fn(strings.TrimSuffix(line, "\n")) {
					break
				}
			}
		}
		return nil
	}
}

// ReadAll .
func ReadAll(fn func(data string) error) reader {
	return func(r io.Reader) error {
		data, _ := io.ReadAll(r)
		return fn(string(data))
	}
}

// WriteTo .
func WriteTo(w io.Writer) reader {
	return func(r io.Reader) error {
		_, err := io.Copy(w, r)
		return err
	}
}

// Unmarshal .
func Unmarshal(v interface{}) reader {
	return func(r io.Reader) error {
		if data, err := io.ReadAll(r); err != nil {
			return errors.Wrap(err, "http")
		} else {
			return errors.Wrap(json.Unmarshal(data, v), "json unmarshal")
		}
	}
}