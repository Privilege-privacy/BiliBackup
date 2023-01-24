package main

import (
	"BiliBackup/core"
	bili "BiliBackup/core/bili"
	"flag"
	"fmt"
	"github.com/monkeyWie/gopeed-core/pkg"
)

var (
	favid    int
	Format   bool
	Connects int
)

func init() {
	core.CheckDownloadDir()
	flag.IntVar(&favid, "f", 0, "收藏夹ID")
	flag.IntVar(&bili.Pagenumber, "pn", 1, "需要备份的页数，不指定时每次备份将备份最新的一页")
	flag.StringVar(&bili.RemotePeth, "remote", "", "Rclone 挂载的驱动名和路径")
	flag.IntVar(&Connects, "n", 3, "下载线程数")
	flag.BoolVar(&bili.Convert, "convert", false, "是否转换视频格式后上传")
}

func main() {
	flag.Parse()
	if favid == 0 {
		fmt.Println("请指定收藏夹ID")
		return
	}
	if bili.Pagenumber == 0 {
		bili.Pagenumber = 1
	}
	if bili.RemotePeth == "" {
		fmt.Println("-remote 没有填写，请指定Rclone挂载的驱动名和路径")
		return
	}

	pkg.Connects = Connects

	bili.GetfavBvid(favid)
}
