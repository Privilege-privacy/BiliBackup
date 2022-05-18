package core

import (
	"crypto/md5"
	"fmt"
	"strings"
)

var _entropy = "rbMCKn@KuamXWlPMoJGsKcbiJKUfkPF_8dABscJntvqhRSETg"
var _paramsTemp = "appkey=%s&cid=%s&otype=json&qn=%s&quality=%s&type="
var _playApiTemp = "https://interface.bilibili.com/v2/playurl?%s&sign=%s"
var _quality = "80"

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
