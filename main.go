package main

import (
	"os"
	"regexp"
	"sync"

	"github.com/charlesbases/progressbar"
	"github.com/urfave/cli"

	"github.com/charlesbases/viper/logger"
	"github.com/charlesbases/viper/website"
	"github.com/charlesbases/viper/website/xvideos"
)

// home 网站首页
var home = regexp.MustCompile(`https://([^/]+)`)

// homeHook 视频网站首页
var homeHook = map[string]func(link website.Link) website.WebHook{
	"www.xvideos.com": xvideos.New,
}

var app = &cli.App{
	Name:      "viper",
	Usage:     "video download tool",
	ArgsUsage: "[link]",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "f",
			Usage: "load video link from file",
			Value: "index.txt",
		},
		cli.IntFlag{
			Name:  "c",
			Usage: "number of videos downloaded at the same time",
			Value: 10,
		},
	},
	Action: func(ctx *cli.Context) error {
		website.SetConcurrent(ctx.Int("c"))

		// with os.args
		if len(ctx.Args()) != 0 {
			return dload(os.Args...)
		}
		// with file
		links, err := website.ReadLineWithFile(ctx.String("f"))
		if err != nil {
			return err
		}
		return dload(links...)
	},
}

// dload .
func dload(links ...string) error {
	reader := progressbar.NewReader()
	defer reader.Close()

	var wg sync.WaitGroup
	wg.Add(len(links))

	for _, link := range links {
		weblink := link
		go func() {
			// 解析链接网站首页
			if webhome := website.FindSubstring(home, weblink); len(webhome) != 0 {
				if hook, found := homeHook[webhome]; found {
					logger.Error(website.H(reader, hook(website.Link(weblink))))
				}
			}
			wg.Done()
		}()
	}

	wg.Wait()
	return nil
}

func main() {
	defer logger.Close()

	if err := app.Run(os.Args); err != nil {
		logger.Error(err)
	}
}
