package core

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"BiliBackup/core"

	"github.com/fatih/color"
	"github.com/glebarez/sqlite"
	"github.com/melbahja/got"
	"github.com/tidwall/gjson"
	"gorm.io/gorm"
)

var (
	Convert          bool
	DB               *gorm.DB
	currentDirectory string
	DownloadPath     string
	maxRetries       int
)

type Bili struct {
	gorm.Model
	Bvid string
}

func init() {
	database, err := gorm.Open(sqlite.Open("bili.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	DB = database

	if err := DB.AutoMigrate(&Bili{}); err != nil {
		log.Fatalf("failed to migrate table: %v", err)
	}

	maxRetries = 3
	currentDirectory, _ := os.Getwd()
	DownloadPath = currentDirectory + "/download/"
}

func getCid(bvid string) (cid string, title string, err error) {
	videoAPI := fmt.Sprintf("https://api.bilibili.com/x/web-interface/view?bvid=%s", bvid)
	body := core.DoGet("GET", videoAPI)

	cid = gjson.Get(body, "data.cid").String()
	title = gjson.Get(body, "data.title").String()

	if cid == "" || title == "" {
		return "", "", fmt.Errorf("failed to get video info")
	}

	if !Convert {
		title += ".flv"
	}

	title = strings.ReplaceAll(title, "/", "")
	return cid, title, nil
}

func getDownloadUrl(bvid string) (url, title string, err error) {
	cid, title, err := getCid(bvid)
	if err != nil || cid == "" || title == "" {
		return "", "", fmt.Errorf("failed to get download URL: %s", bvid)
	}
	body := core.DoGet("GET", core.GenGetAidChildrenParseFun(cid))
	return gjson.Get(body, "durl.0.url").String(), title, nil
}

func download(url string, downloadPath string) error {
	download := got.NewDownload(context.Background(), url, downloadPath)
	download.Header = []got.GotHeader{
		{"User-Agent", core.UserAgent},
		{"Referer", core.Referer},
	}

	if err := download.Init(); err != nil {
		log.Printf("Failed to initialize download: %v", err)
		return err
	}

	if err := download.Start(); err != nil {
		log.Printf("Failed to start download: %v", err)
		return err
	}

	return nil
}

func downloadVideoSingleThreaded(url string, filename string) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", core.UserAgent)
	req.Header.Set("Referer", core.Referer)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	done := make(chan error)
	go func() {
		_, err = io.Copy(file, resp.Body)
		done <- err
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for retries := 0; retries < 3; retries++ {
		select {
		case err := <-done:
			if err != nil {
				return err
			}
			return nil
		case <-ticker.C:
			go func() {
				resp, err = client.Do(req)
				if err != nil {
					done <- err
					return
				}
				defer resp.Body.Close()
				done <- nil
			}()
		}
	}
	return fmt.Errorf("failed to download video after 3 retries")
}

func Run(videoID, cloudDrivePath string) {
	database := DB
	if database.Where("bvid = ?", videoID).Limit(1).Find(&Bili{}).RowsAffected == 1 {
		return
	}

	videoURL, videoTitle, err := getDownloadUrl(videoID)
	if err != nil {
		log.Println(err)
		return
	}

	VideoPath := filepath.Join(DownloadPath, videoTitle)

	var downloaded bool
	for retries := 0; retries < maxRetries && !downloaded; retries++ {
		downloaded, err = downloadVideoMultiThreaded(videoURL, VideoPath)
		if err != nil {
			log.Printf("download failed (%d/%d): %s, retrying... (error: %v)\n", retries+1, maxRetries, videoTitle, err)
			cleanupDownload(VideoPath)
		}
	}

	if !downloaded {
		log.Printf("multi-threaded download of %s failed after %d retries, attempting single-threaded download...\n", videoTitle, maxRetries)
		if err := downloadVideoSingleThreaded(videoURL, VideoPath); err != nil {
			log.Printf("single-threaded download of %s failed: (%v)\n", videoTitle, err)
			return
		}
	}

	if Convert {
		convertVideo(VideoPath)
	}

	log.Println("uploading:", videoTitle)
	if err := uploadVideo(cloudDrivePath, DownloadPath); err != nil {
		log.Println("could not upload file:", err)
		return
	}

	if err := database.Create(&Bili{Bvid: videoID}).Error; err != nil {
		log.Println("could not save videoID to database:", err)
	}
	color.Yellow("-------------")
}

func downloadVideoMultiThreaded(url, title string) (bool, error) {
	if err := download(url, title); err != nil {
		return false, err
	}
	return true, nil
}

func cleanupDownload(VideoPath string) {
	if Convert {
		os.Remove(VideoPath + ".mp4")
	} else {
		os.Remove(VideoPath + ".flv")
	}
}

func uploadVideo(remote, local string) error {
	url := "http://127.0.0.1:5572/sync/copy?srcFs=" + local + "&dstFs=" + remote + "&createEmptySrcDirs=true"
	body := core.DoGet("POST", url)
	if strings.Contains(body, "{}") {
		return nil
	}
	return fmt.Errorf("上传失败: (%s)", local)
}

func convertVideo(filename string) error {
	fmt.Printf("正在转换 %s 视频格式...\n", filename)
	cmd := exec.Command("ffmpeg", "-i", filename, "-c:v", "copy", "-c:a", "copy", filename+".mp4")
	if err := cmd.Run(); err != nil {
		return err
	}
	if err := os.Remove(filename); err != nil {
		return err
	}
	return nil
}
