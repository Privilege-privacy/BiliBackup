package core

import (
	"fmt"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"log"
	"net/http"
)

func GetfavBvid(favid, pagenumber int, remotepeth string) {
loop:
	for i := 1; i < pagenumber; i++ {
		if err := func(i int) bool {
			url := fmt.Sprintf("https://api.bilibili.com/x/v3/fav/resource/list?media_id=%d&pn=%d&ps=20&keyword=&order=mtime&type=0&tid=0&platform=web&jsonp=jsonp", favid, i)
			client := http.Client{}
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				log.Println(err)
			}

			resp, err := client.Do(req)

			if err != nil {
				log.Println(err)
			}

			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Println(err)
			}

			bvid := gjson.Get(string(body), "data.medias.#.bvid")
			bvid.ForEach(func(key, value gjson.Result) bool {
				//下载并上传视频
				Run(value.String(), remotepeth)
				return true
			})

			if hasMore := gjson.Get(string(body), "data.has_more"); hasMore.String() == "false" {
				return true
			}
			return false
		}(i); err {
			break loop
		}
	}
}
