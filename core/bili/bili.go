package core

import (
	"fmt"

	"github.com/tidwall/gjson"
)

var (
	Pagenumber int
	RemotePath string
	Thread     int64
)

func GetFavoriteBVIDs(favoriteID int) {
	for pageNumber := 0; pageNumber <= Pagenumber; pageNumber++ {
		body, _ := Client.R().Get(fmt.Sprintf("https://api.bilibili.com/x/v3/fav/resource/list?media_id=%d&pn=%d&ps=20&keyword=&order=mtime&type=0&tid=0&platform=web&jsonp=jsonp", favoriteID, pageNumber))

		gjson.Get(body.String(), "data.medias").ForEach(func(key, value gjson.Result) bool {
			if value.Get("title").String() == "已失效视频" {
				return true
			}

			NewVideo(value.Get("bvid").String(), Thread).Run()
			return true
		})

		if hasMore := gjson.Get(body.String(), "data.has_more"); hasMore.String() == "false" {
			break
		}
	}
}
