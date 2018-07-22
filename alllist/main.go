package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"golang.org/x/net/html"
	iconv "gopkg.in/iconv.v1"
)

const (
	url string = "http://quote.eastmoney.com/stocklist.html"
)

func main() {
	chars, err := FetchPageGBK(url)
	if err != nil {
		return
	}

	ParseHtml(bytes.NewReader(chars))
}

func FetchPageGBK(url string) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Cant create Request obj: ", err)
		return nil, err
	}
	//req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	//req.Header.Add("Accept-Encoding", "gzip, deflate")
	//req.Header.Add("Accept-Language", "zh-cn,zh;q=0.8,en-us;q=0.5,en;q=0.3")
	//req.Header.Add("Connection", "keep-alive")
	//req.Header.Add("Host", uri.Host)
	//req.Header.Add("Referer", uri.String())
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:12.0) Gecko/20100101 Firefox/12.0")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Req err!:", err)
		return nil, err
	}
	defer resp.Body.Close()
	// reader := transform.NewReader(resp.Body, simplifiedchinese.HZGB2312.NewDecoder())
	// buf := bytes.NewBuffer(make([]byte, 1024))
	// io.Copy(buf, reader)

	// convert resp.Body from gbk to utf-8 format
	cd, err := iconv.Open("utf-8", "gbk")
	if err != nil {
		fmt.Println("iconv.Open failed!")
		return nil, err
	}
	defer cd.Close()

	var bf bytes.Buffer
	bufsize := 512
	r := iconv.NewReader(cd, resp.Body, bufsize)
	_, err = io.Copy(&bf, r)
	if err != nil {
		fmt.Println("\nio.Copy failed in fetchpage: error code: %s", err)
		io.Copy(os.Stdin, resp.Body)
		os.Exit(1)
		return nil, err
	}

	return bf.Bytes(), nil
}

func ParseHtml(r io.Reader) {
	doc, err := html.Parse(r)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// 寻找节点树<div class="quotebody"> ... </div>
	var quotebody *html.Node
	var f func(n *html.Node) bool
	f = func(n *html.Node) bool {
		if n.Type == html.ElementNode && n.Data == "div" {
			for _, a := range n.Attr {
				if a.Key == "class" && a.Val == "quotebody" {
					quotebody = n
					return true
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if f(c) {
				return true
			}
		}
		return false
	}
	if f(doc) == false {
		fmt.Println("Not find quotebody, and Exit!")
		os.Exit(1)
	}

	// 将quotebody下的所有"li"元素收集起来，不需要先查找"ul"元素，因为只有股票代码放置在"li"内
	var allli []string
	var g func(n *html.Node) // 闭包，收集所有"ul"元素到allul
	g = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "li" {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if c.Type == html.ElementNode && c.Data == "a" && c.Attr[0].Key == "target" {
					allli = append(allli, c.FirstChild.Data)
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			g(c)
		}
	}
	if g(quotebody); len(allli) == 0 {
		fmt.Println("No li in quotebody")
		os.Exit(1)
	}

	for _, li := range allli {
		//	fmt.Fprintln(os.Stdout, li)
		j0 := strings.Index(li, "(")
		j1 := strings.Index(li, ")")
		if j0 == j1 || j0 == -1 || j1 == -1 {
			continue
		}
		fmt.Fprintf(os.Stdout, "%s, %s\n", li[:j0], li[j0+1:j1])
	}
}
