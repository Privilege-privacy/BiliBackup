package core

import (
	"fmt"

	"BiliBackup/core"

	"github.com/tidwall/gjson"
)

var (
	Pagenumber int
	RemotePath string
)

func GetFavoriteBVIDs(favoriteID int) {
	for pageNumber := 0; pageNumber <= Pagenumber; pageNumber++ {
		body := core.DoGet("GET", fmt.Sprintf("https://api.bilibili.com/x/v3/fav/resource/list?media_id=%d&pn=%d&ps=20&keyword=&order=mtime&type=0&tid=0&platform=web&jsonp=jsonp", favoriteID, pageNumber))

		gjson.Get(body, "data.medias").ForEach(func(key, value gjson.Result) bool {
			if value.Get("title").String() == "已失效视频" {
				return true
			}
			Run(value.Get("bvid").String(), RemotePath)
			return true
		})

		if hasMore := gjson.Get(body, "data.has_more"); hasMore.String() == "false" {
			break
		}
	}
}
