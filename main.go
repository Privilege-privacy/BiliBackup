package main

import (
	"flag"
	"fmt"

	"BiliBackup/core"
	bili "BiliBackup/core/bili"
)

var (
	favid    int
	Format   bool
	Connects int
)

func init() {
	core.CheckDownloadDir()
	flag.IntVar(&favid, "f", 0, "收藏夹ID")
	flag.IntVar(&bili.Pagenumber, "pn", 100000, "默认备份整个收藏夹的视频，可以指定备份的页数")
	flag.StringVar(&bili.RemotePath, "remote", "", "Rclone 挂载的驱动名和路径")
	flag.BoolVar(&bili.Convert, "convert", false, "是否转换视频格式后上传")
}

func main() {
	flag.Parse()
	if favid == 0 {
		fmt.Println("请指定收藏夹ID")
		return
	}
	if bili.RemotePath == "" {
		fmt.Println("-remote 没有填写，请指定Rclone挂载的驱动名和路径")
		return
	}

	bili.GetFavoriteBVIDs(favid)
}
