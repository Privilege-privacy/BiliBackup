package core

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"BiliBackup/core"

	"github.com/fatih/color"
	"github.com/glebarez/sqlite"
	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	Convert      bool
	DB           *gorm.DB
	Client       *resty.Client
	DownloadPath string
)

func init() {
	database, err := gorm.Open(sqlite.Open("bili.db"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	DB = database

	if err := DB.AutoMigrate(&Bili{}); err != nil {
		log.Fatalf("failed to migrate table: %v", err)
	}

	Client = resty.New()
	Client.SetRetryCount(3)

	currentDirectory, _ := os.Getwd()
	DownloadPath = currentDirectory + "/download/"

	core.CheckDownloadDir()
}

type Bili struct {
	gorm.Model
	Bvid string
}

type VideoInfo struct {
	bvid     string
	cid      string
	aid      string
	download *core.Downloader
}

func NewVideo(bvid string, p int64) *VideoInfo {
	return &VideoInfo{
		bvid: bvid,
		download: &core.Downloader{
			Process: p,
			Headers: []*core.Header{
				{
					Key:   "Referer",
					Value: "https://www.bilibili.com/",
				},
				{
					Key:   "User-Agent",
					Value: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/83.0.4103.116 Safari/537.36",
				},
			},
			Client: &http.Client{},
		},
	}
}

func (v *VideoInfo) GetVideoInfo() (err error) {
	resp, err := Client.R().Get(fmt.Sprintf("https://api.bilibili.com/x/web-interface/view?bvid=%s", v.bvid))
	if err != nil {
		return err
	}

	v.cid = gjson.Get(resp.String(), "data.cid").String()
	v.aid = gjson.Get(resp.String(), "data.aid").String()

	v.download.FileName += gjson.Get(resp.String(), "data.title").String()
	v.download.FileName = regexp.MustCompile(`[<>:"/\\|?*\x00-\x1F\x7F]+`).ReplaceAllString(v.download.FileName, "")
	v.download.FileName = filepath.Join(DownloadPath, v.download.FileName)

	if !Convert {
		v.download.FileName += ".flv"
	}

	return v.getWebDownloadURL()
}

func (v *VideoInfo) getWebDownloadURL() (err error) {
	resp, err := Client.R().Get(core.GenGetAidChildrenParseFun(v.cid))
	v.download.Url = gjson.Get(resp.String(), "durl.0.url").String()
	return err
}

func (v *VideoInfo) getTVDownloadURL() (audio string, video string) {
	resp, _ := Client.R().Get(fmt.Sprintf("https://api.snm0516.aisee.tv/x/tv/ugc/playurl?avid=%s&mobi_app=android_tv_yst&fnval=80&qn=120&cid=%s&access_key=&fourk=1&platform=android&device=android&build=103800&fnver=0", v.aid, v.cid))
	return gjson.Get(resp.String(), "dash.audio.0.base_url").String(), gjson.Get(resp.String(), "dash.video.0.base_url").String()
}

func (v *VideoInfo) mergeAudioAndVideo() error {
	audio, video := v.getTVDownloadURL()

	color.Red("无法从 Web 端获取下载地址，使用 Tv 端接口分别下载音视频，默认不执行音视频合并")
	videoName := v.download.FileName + ".flv"
	audioName := v.download.FileName + ".mp3"

	if err := core.NewDownloader(audio, audioName, v.download.Process).Start(); err != nil {
		return err
	}
	if err := core.NewDownloader(video, videoName, v.download.Process).Start(); err != nil {
		return err
	}

	if Convert {
		if err := exec.Command("ffmpeg", "-i", videoName,
			"-i", audioName, "-c:v", "copy", "-c:a", "aac",
			v.download.FileName+".mp4").Run(); err != nil {
			return err
		}
	}

	os.Remove(audioName)
	os.Remove(videoName)

	return v.upload()
}

func (v *VideoInfo) Run() (err error) {
	database := DB
	if database.Where("bvid = ?", v.bvid).Limit(1).Find(&Bili{}).RowsAffected == 1 {
		return
	}

	if err = v.GetVideoInfo(); err != nil {
		return errors.Wrap(err, "获取视频信息失败: ")
	}

	if v.download.Url == "" {
		return v.mergeAudioAndVideo()
	}

	if err := v.download.Start(); err != nil {
		return err
	}
	color.Green("下载完成：%s\n", v.download.FileName)

	if Convert {
		if err := exec.Command("ffmpeg", "-i", v.download.FileName,
			"-c:v", "libx264", "-c:a", "aac",
			v.download.FileName+".mp4").Run(); os.Remove(v.download.FileName) != nil {
			return err
		}
		v.download.FileName += ".mp4"
	}
	return v.upload()
}

func (v *VideoInfo) upload() (err error) {
	resp, err := Client.R().Post("http://127.0.0.1:5572/sync/copy?srcFs=" + DownloadPath + "&dstFs=" + RemotePath + "&createEmptySrcDirs=true")
	if err != nil {
		return err
	}
	defer os.Remove(v.download.FileName)
	if strings.Contains(resp.String(), "{}") {
		color.Green("上传完成：%s\n", v.download.FileName)

		if err := DB.Create(&Bili{Bvid: v.bvid}).Error; err != nil {
			errors.Wrap(err, "保存到数据库失败: %v")
		}

		return nil
	}
	return
}
