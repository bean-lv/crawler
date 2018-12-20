package finder

import (
	"flag"
	"fmt"
	"os"
)

// 命令参数。
var (
	firstURL string
	domains  string
	depth    uint
	dirPath  string
)

func init() {
	flag.StringVar(&firstURL, "first", "http://zhihu.sogou.com/zhihu?query=golang+logo",
		"The first URL which you want to access.")
	flag.StringVar(&domains, "domains", "zhihu.com",
		"The primary domains which you accepted. "+
			"Please using comma-separated multiple domains.")
	flag.UintVar(&depth, "depth", 3,
		"The depth for crawling.")
	flag.StringVar(&dirPath, "dir", "./pictures",
		"The path which you want to save the image files.")
}

func Usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "\tfinder [flags] \n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
}
