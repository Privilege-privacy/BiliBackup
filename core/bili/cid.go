package core

import (
	"BiliBackup/core"
	"fmt"
	"github.com/glebarez/sqlite"
	"github.com/monkeyWie/gopeed-core/pkg"
	"github.com/tidwall/gjson"
	"gorm.io/gorm"
	"log"
	"os"
	"os/exec"
)

var (
	Convert bool
	DB      *gorm.DB
	curdir  string
)

type Bili struct {
	gorm.Model
	Bvid string
}

func init() {
	db, err := gorm.Open(sqlite.Open("bili.db"), &gorm.Config{})
	DB = db
	if err == nil {
		DB.AutoMigrate(&Bili{})
	}

	curdir, _ = os.Getwd()
	pkg.DownloadPath = curdir + "/download"
}

func getCid(bvid string) (cid string, tittle string, err error) {
	body := core.DoGet("GET", "https://api.bilibili.com/x/web-interface/view?bvid="+bvid)

	cid = gjson.Get(body, "data.cid").String()
	tittle = gjson.Get(body, "data.title").String()

	if !Convert {
		tittle += ".flv"
	}

	return cid, tittle, nil
}

func getDownloadUrl(bvid string) (url, title string, err error) {
	cid, tittle, err := getCid(bvid)
	body := core.DoGet("GET", core.GenGetAidChildrenParseFun(cid))

	return gjson.Get(body, "durl.0.url").String(), tittle, err
}

func Run(bvid, cloudDrivePath string) {

	if DB.Where("Bvid = ?", bvid).Limit(1).Find(&Bili{}).RowsAffected == 1 {
		return
	}

	url, title, err := getDownloadUrl(bvid)

	if err == nil {
		log.Println("开始下载： ", title)
		pkg.Download(url, title, map[string]string{
			"User-Agent": core.UserAgent,
			"Referer":    core.Referer,
		})

		DB.Create(&Bili{Bvid: bvid})

		if Convert {
			formatConversion(title)
		}

		log.Println("正在上传文件...")
		upload(cloudDrivePath, curdir+"/download/")

		if Convert {
			os.Remove(curdir + "/download/" + title + ".mp4")
		} else {
			os.Remove(curdir + "/download/" + title)
		}

	}
}

func upload(remote, local string) {
	url := "http://127.0.0.1:5572/sync/copy?srcFs=" + local + "&dstFs=" + remote + "&createEmptySrcDirs=true"
	body := core.DoGet("POST", url)
	if body == "{}" {
		fmt.Println("upload success")
	}
}

func formatConversion(filename string) error {
	fmt.Printf("正在转换 %s 视频格式...\n", filename)
	exec.Command("ffmpeg", "-i", "download/"+filename, "-c:v", "copy", "-c:a", "copy", "download/"+filename+".mp4").Run()
	os.Remove("download/" + filename)
	return nil
}
