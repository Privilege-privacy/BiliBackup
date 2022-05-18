package core

import (
	"fmt"
	"github.com/schollz/progressbar/v3"
	"github.com/tidwall/gjson"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

var userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.149 Safari/537.36"

func getCid(bvid string) (cid string, tittle string, err error) {
	var (
		cidUrl string
	)
	cidUrl = "https://api.bilibili.com/x/web-interface/view?bvid=" + bvid
	client := http.Client{}
	req, err := http.NewRequest("GET", cidUrl, nil)
	if err != nil {
		return
	}
	req.Header.Set("User-Agent", userAgent)
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	cid = gjson.Get(string(body), "data.cid").String()
	tittle = gjson.Get(string(body), "data.title").String()
	//println(cid.String(), tittle.String())
	return cid, tittle, nil
}

func getDownloadUrl(cid string) (url string, err error) {
	url = GenGetAidChildrenParseFun(cid)
	client := http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	req.Header.Set("User-Agent", userAgent)
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	url = gjson.Get(string(body), "durl.0.url").String()
	return url, nil
}

func download(url string, fileName string) (err error) {
	client := http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	req.Header.Set("User-Agent", userAgent)
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	file, err := os.Create(filepath.Join("download", fileName))
	if err != nil {
		return
	}
	defer file.Close()
	fmt.Println("Downloading", fileName)
	bar := progressbar.DefaultBytes(
		resp.ContentLength,
		"Downloading...",
	)
	io.Copy(io.MultiWriter(file, bar), resp.Body)
	return nil
}

func Run(bvid, cloudDrivePath string) {
	cid, tittle, err := getCid(bvid)
	if err != nil {
		log.Println(err)
	}
	url, err := getDownloadUrl(cid)
	if err != nil {
		log.Println(err)
	}
	err = download(url, tittle+".flv")
	if err != nil {
		log.Println(err)
	}
	err = formatConversion(tittle)
	if err != nil {
		log.Println("视频转换格式时错误", err)
	}
	curdir, _ := os.Getwd()
	log.Println("正在上传文件...")
	upload(cloudDrivePath, curdir+"/download/")
	defer os.Remove(curdir + "/download/" + tittle + ".mp4")
}

func upload(remote, local string) {
	url := "http://127.0.0.1:5572/sync/copy?srcFs=" + local + "&dstFs=" + remote + "&createEmptySrcDirs=true"
	client := http.Client{}
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		log.Println("访问 Rclone rcd 接口错误", err)
	}
	res, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
	}
	if res.StatusCode == 200 || string(body) == "{}" {
		fmt.Println("upload success")
	}
}

func formatConversion(filename string) error {
	fmt.Printf("正在转换 %s 视频格式...\n", filename)
	exec.Command("ffmpeg", "-i", "download/"+filename+".flv", "-c:v", "copy", "-c:a", "copy", "download/"+filename+".mp4").Run()
	os.Remove("download/" + filename + ".flv")
	return nil
}
