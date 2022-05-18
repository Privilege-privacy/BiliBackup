package main

import (
	"BiliBackup/core"
	"flag"
	"fmt"
)

var (
	favid      *int
	pagenumber *int
	remotepath *string
)

func init() {
	core.CheckDownloadDir()
	favid = flag.Int("f", 0, "收藏夹ID")
	pagenumber = flag.Int("pn", 1, "需要备份的页数，不指定时每次备份将备份最新的一页")
	remotepath = flag.String("remote", "", "Rclone 挂载的驱动名和路径")
}

func main() {
	flag.Parse()
	if *favid == 0 {
		fmt.Println("请指定收藏夹ID")
		return
	}
	if *pagenumber == 0 {
		*pagenumber = 1
	}
	if *remotepath == "" {
		fmt.Println("-remote 没有填写，请指定Rclone挂载的驱动名和路径")
		return
	}
	core.GetfavBvid(*favid, *pagenumber, *remotepath)
}
