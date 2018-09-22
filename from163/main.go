package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

const (
	url string = "http://quotes.money.163.com/service/chddata.html?code=0600000&start=20140101&end=20151231"
)

func main() {
	fmt.Println("vim-go")
}

func FetchPage(url string, gb2312 bool) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Cant create Request obj: ", err)
		return nil, err
	}
	//req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	//req.Header.Set("Accept-Encoding", "gzip, deflate")
	//req.Header.Set("Accept-Language", "zh-cn,zh;q=0.8,en-us;q=0.5,en;q=0.3")
	//req.Header.Set("Connection", "keep-alive")
	//req.Header.Set("Host", uri.Host)
	//req.Header.Set("Referer", uri.String())
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:12.0) Gecko/20100101 Firefox/12.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("http.Get's status code is not %d", 200)
	}
	defer resp.Body.Close()

	var rd io.Reader
	if gb2312 { // convert gb2312 into utf8
		rd = transform.NewReader(resp.Body, simplifiedchinese.GBK.NewDecoder())
	} else {
		rd = resp.Body
	}
	return ioutil.ReadAll(rd)
}
