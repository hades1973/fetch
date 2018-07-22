package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	iconv "gopkg.in/iconv.v1"
)

type StockBonus struct {
	Year  string  // 分红年度
	Base  float64 // 一般为10股
	Song  float64 // 每Base股送多少股
	Zeng  float64 // 每Base股转增多少股
	Pai   float64 // 派息(元)
	GQDJR string  // 股权登记日
	GQJZR string  // 股权基准日
	HGSSR string  // 红股上市日
}

const (
	prefixUrl string = `http://www.cninfo.com.cn/information/dividend/szmb%s.html`
)

func main() {
	//stocks := []string{
	//	"000540",
	//	"000541",
	//}

	f, err := os.Open("000540.html")
	if err != nil {
		log.Fatal(err)
		return
	}
	defer f.Close()

	records, err := QueryHTML(f)
	if err != nil {
		log.Fatal(err)
	}
	// rec:= records[i]各下标元素内容、意义如下:
	// i: 意义：		举例
	// 0: 分红年度		2014年度
	// 1: 分红方案		10送5转增10股派2元(含税)
	// 2: 股权登记日	20150331
	// 3: 除权基准日	20150401
	// 4: 红股上市日	20150401
	fmt.Println("年度\t股基\t送股\t转增\t派息\t登记日\t基准日\t上市日\n")
	var rsb StockBonus
	var base, song, zhuan, pai string
	for _, rec := range records {
		base, song, zhuan, pai = "", "", "", ""
		r, _ := regexp.Compile("\\d+(\\.\\d+)?")
		s := rec[1]
		base = r.FindString(s)
		if i := strings.Index(s, "送"); i > 0 {
			song = r.FindString(s[i:])
		}
		if j := strings.Index(s, "转增"); j > 0 {
			zhuan = r.FindString(s[j:])
		}
		if k := strings.Index(s, "派"); k > 0 {
			pai = r.FindString(s[k:])
		} else {
			base = ""
		}
		rsb.Year = r.FindString(rec[0])
		rsb.Base, _ = strconv.ParseFloat(base, 64)
		rsb.Song, _ = strconv.ParseFloat(song, 64)
		rsb.Zeng, _ = strconv.ParseFloat(zhuan, 64)
		rsb.Pai, _ = strconv.ParseFloat(pai, 64)
		rsb.GQDJR = r.FindString(rec[2])
		rsb.GQJZR = r.FindString(rec[3])
		rsb.HGSSR = r.FindString(rec[4])
		fmt.Printf("%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\n",
			rsb.Year,
			rsb.Base,
			rsb.Song,
			rsb.Zeng,
			rsb.Pai,
			rsb.GQDJR,
			rsb.GQJZR,
			rsb.HGSSR,
		)
	}

	return

}

func FetchPage(url string) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Cant create Request obj: ", err)
		return nil, err
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:12.0) Gecko/20100101 Firefox/12.0")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Req err!:", err)
		return nil, err
	}
	defer resp.Body.Close()

	buf := bytes.NewBuffer(make([]byte, 0))
	io.Copy(buf, resp.Body)
	return buf.Bytes(), nil
}

func ConvertGBKToUTF8(chars []byte) ([]byte, error) {
	// convert resp.Body from gbk to utf-8 format
	cd, err := iconv.Open("utf-8", "gbk")
	if err != nil {
		fmt.Println("iconv.Open failed!")
		return nil, err
	}
	defer cd.Close()

	var buf bytes.Buffer
	bufsize := 512
	r := iconv.NewReader(cd, bytes.NewReader(chars), bufsize)
	_, err = io.Copy(&buf, r)

	return buf.Bytes(), nil
}

func QueryHTML(r io.Reader) ([][]string, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, errors.New("goquery.NewDocumentFromReader can't create doc!")
	}
	// find <div class="clear" ...>
	//        <table ...>
	//          <tr>
	// 内所有tr元素
	var records = [][]string{}
	if trs := doc.Find(".clear table tr"); trs != nil {
		trs.Each(func(i int, n *goquery.Selection) {
			if i == 0 { // skip head
				return
			}
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
