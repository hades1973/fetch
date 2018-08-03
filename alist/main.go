package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

const (
	url string = "http://query.sse.com.cn/security/stock/downloadStockListFile.do?csrcCode=&stockCode=&areaName=&stockType=1"
)

func main() {
	// 打印网页内容
	chars, err := FetchPage(url, true)
	if err != nil {
		log.Fatal(err)
		return
	}
	fmt.Println(string(chars))
}

func FetchPage(url string, gb2302 bool) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Cant create Request obj: ", err)
		return nil, err
	}
	req.Header.Set("User-Agent",
		"Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.108 Safari/537.36")
	req.Header.Set("Host", "query.sse.com.cn")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Origin", "http://www.sse.com.cn")
	//req.Header.Set("Referer", "http://www.sse.com.cn/assortment/stock/list/share")
	req.Header.Set("Accept-Encoding", "gzip,deflate")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.Status[0] != 2 {
		return nil, errors.New("Request not success, Status: " + resp.Status)
	}
	defer resp.Body.Close()

	if gb2302 { // convert gb2312 into utf8
		rdGBK := transform.NewReader(resp.Body, simplifiedchinese.GBK.NewDecoder())
		chars, err := ioutil.ReadAll(rdGBK)
		if err != nil {
			return nil, err
		}
		return chars, nil
	}

	// default code is utf8
	buf := bytes.NewBuffer(nil)
	io.Copy(buf, resp.Body)
	return buf.Bytes(), nil
}
