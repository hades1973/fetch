package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func main() {
	// 打印网页内容
	url := os.Args[1]
	chars, err := FetchPage(url)
	if err != nil {
		log.Fatal(err)
		return
	}
	fmt.Println(string(chars))
}

func FetchPage(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.108 Safari/537.36")
	// req.Header.Set("Host", "query.sse.com.cn")
	// req.Header.Set("Connection", "keep-alive")
	// req.Header.Set("Accept", "*/*")
	// req.Header.Set("Origin", "http://www.sse.com.cn")
	// req.Header.Set("Referer", "http://www.sse.com.cn/assortment/stock/list/share")
	// req.Header.Set("Accept-Encoding", "gzip,deflate")
	// req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.Status[0] != '2' {
		return nil, errors.New(resp.Status)
	}
	defer resp.Body.Close()
	buf := bytes.NewBuffer(nil)
	io.Copy(buf, resp.Body)
	return buf.Bytes(), nil
}
