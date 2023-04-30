package core

import (
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"strings"
)

var (
	_entropy     = "rbMCKn@KuamXWlPMoJGsKcbiJKUfkPF_8dABscJntvqhRSETg"
	_paramsTemp  = "appkey=%s&cid=%s&otype=json&qn=%s&quality=%s&type="
	_playApiTemp = "https://interface.bilibili.com/v2/playurl?%s&sign=%s"
	_quality     = "80"
	UserAgent    = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.149 Safari/537.36"
	Referer      = "https://www.bilibili.com/"
)

func GenGetAidChildrenParseFun(cid string) (urlApi string) {
	appKey, sec := GetAppKey(_entropy)
	params := fmt.Sprintf(_paramsTemp, appKey, cid, _quality, _quality)
	chksum := fmt.Sprintf("%x", md5.Sum([]byte(params+sec)))
	urlApi = fmt.Sprintf(_playApiTemp, params, chksum)
	return urlApi
}

func GetAppKey(entropy string) (appkey, sec string) {
	revEntropy := ReverseRunes([]rune(entropy))
	for i := range revEntropy {
		revEntropy[i] = revEntropy[i] + 2
	}
	ret := strings.Split(string(revEntropy), ":")

	return ret[0], ret[1]
}

func ReverseRunes(runes []rune) []rune {
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	return runes
}

func DoGet(method, url string) string {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return ""
	}
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Referer", Referer)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err == nil {
		return string(body)
	}
	return ""
}
