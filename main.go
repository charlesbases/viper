package main

import (
	"os"
	"regexp"

	"github.com/urfave/cli"

	"github.com/charlesbases/viper/logger"
	"github.com/charlesbases/viper/website"
	"github.com/charlesbases/viper/website/xvideos"
)

// home 网站首页
var home = regexp.MustCompile(`https://([^/]+)`)

// homeHook 视频网站首页
var homeHook = map[string]func(link website.Link) website.WebHook{
	"www.xvideos.com": xvideos.H,
}

var app = &cli.App{
	Name:     "viper",
	Usage:    "video download tool",
	Commands: []cli.Command{},
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "f",
			Usage: "load video link from file",
			Value: "index.txt",
		},
	},
	Action: func(ctx *cli.Context) error {
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

func main() {
	defer logger.Close()

	if err := app.Run(os.Args); err != nil {
		logger.Error(err)
	}
}

// dload .
func dload(links ...string) error {
	for _, link := range links {
		// 解析链接网站首页
		if dst := home.FindStringSubmatch(link); len(dst) > 1 {
			// HomeHook
			if hook, found := homeHook[dst[1]]; found {
				logger.Error(website.H(hook(website.Link(link))))
			}
		}
	}
	return nil
}
