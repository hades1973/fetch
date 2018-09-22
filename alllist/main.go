package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"

	"golang.org/x/net/html"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

const (
	url string = "http://quote.eastmoney.com/stocklist.html"
)

var db *sql.DB

func main() {
	chars, err := FetchPage(url, true)
	if err != nil {
		log.Fatal(err)
	}
	db, _ = sql.Open("mysql", "stockadmin:1973admin@tcp(127.0.0.1:3306)/stockdb")
	defer db.Close()

	ParseHtml(bytes.NewReader(chars))
}

func FetchPage(url string, gb2312 bool) ([]byte, error) {
	client := &http.Client{}
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

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Req err!:", err)
		return nil, err
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

	statement := "insert into stocklist (stockid, stockname) values(?,?)"
	stmt, err := db.Prepare(statement)
	if err != nil {
		log.Fatalln(err)
	}
	defer stmt.Close()
	for _, li := range allli {
		//	fmt.Fprintln(os.Stdout, li)
		j0 := strings.Index(li, "(")
		j1 := strings.Index(li, ")")
		if j0 == j1 || j0 == -1 || j1 == -1 {
			continue
		}
		//fmt.Fprintf(os.Stdout, "%s,%s\n", li[:j0], li[j0+1:j1])
		stockid, stockname := li[j0+1:j1], li[:j0]
		if stockid[:2] == "60" || stockid[:3] == "000" {
			stmt.QueryRow(stockid, stockname)
		}
	}
}
