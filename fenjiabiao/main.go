package main

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

const (
	prefixUrl string = "http://market.finance.sina.com.cn/pricehis.php?symbol=sz%s&startdate=2018-07-11&enddate=2018-07-18"
)

func main() {
	stocks := []string{
		//	"000540",
		"000541",
	}
	for _, stock := range stocks {
		url := fmt.Sprintf(prefixUrl, stock)
		fmt.Println(url)
		chars, err := FetchPage(url)
		if err != nil {
			log.Fatal(err)
		}

		//ioutil.WriteFile(stock+".html", chars, 0776)
		f, err := os.Create(stock + ".csv")
		if err != nil {
			log.Fatalln(err)
		}
		csvWR := csv.NewWriter(f)
		csvWR.Comma = '\t'
		records, err := QueryHTML(bytes.NewReader(chars))
		if err != nil {
			log.Fatalln(err)
		}
		for _, rec := range records {
			csvWR.Write(rec)
		}
		csvWR.Flush()
		f.Close()
	}

	return
}

func FetchPage(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Cant create Request obj: ", err)
		return nil, err
	}
	req.Header.Add("User-Agent",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/62.0.3202.62 Safari/537.36")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Req err!:", err)
		return nil, err
	}
	defer resp.Body.Close()

	// convert gb2312 into utf8
	rdGBK := transform.NewReader(resp.Body, simplifiedchinese.GBK.NewDecoder())
	chars, err := ioutil.ReadAll(rdGBK)
	if err != nil {
		return nil, err
	}
	return chars, nil
}

func QueryHTML(r io.Reader) ([][]string, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, errors.New("goquery.NewDocumentFromReader can't create doc!")
	}
	// find <table id="datalist" ...>
	//        <tbody>
	//          <tr>
	// 内所有tr元素
	var records = [][]string{}
	if trs := doc.Find("#datalist tbody tr"); trs != nil {
		trs.Each(func(i int, n *goquery.Selection) {
			record := []string{}
			n.Find("td").Each(func(j int, c *goquery.Selection) {
				txt, _ := c.Html()
				record = append(record, strings.TrimSpace(txt))
			})
			records = append(records, record)
		})
	}
	return records, nil
}
