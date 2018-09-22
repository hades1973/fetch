package main

import (
	"bufio"
	"bytes"
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	_ "github.com/go-sql-driver/mysql"
	"github.com/tealeg/xlsx"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

const (
	urlShenZheng string = "http://www.szse.cn/api/report/ShowReport?SHOWTYPE=xlsx&CATALOGID=1110&TABKEY=tab1&random=%.17f"
	urlShangHai  string = "http://query.sse.com.cn/security/stock/downloadStockListFile.do?csrcCode=&stockCode=&areaName=&stockType=1"
)

func main() {
	// 从网站拉取股票列表数据，存储到对应的文件里
	{ // 获取深圳股票列表，存储到文件"深股列表.csv"
		// 头文件用chrome或edge使用开发者工具拷贝、黏贴得到
		rand.Seed(time.Now().Unix())
		randX := rand.Float64()
		path := fmt.Sprintf(urlShenZheng, randX)
		chars, err := FetchPage(path, "utf8", "./config/深市.header.csv")
		if err != nil {
			log.Fatalln(err)
			return
		}
		f, _ := os.Create("深股列表.csv")
		outputf := func(s string) {
			f.WriteString(s)
		}
		generateCSVFromXLSXBytes(chars, 0, "\t", outputf)
		f.Close()
	}
	{ // 获取上海股票列表，存储到文件"沪股列表.csv"
		chars, err := FetchPage(urlShangHai, "gb2312", "./config/沪市.header.csv")
		if err != nil {
			log.Fatalln(err)
			return
		}
		rd := csv.NewReader(strings.NewReader(string(chars)))
		rd.Comma = '\t'
		records, err := rd.ReadAll()
		if err != nil {
			log.Fatalln(err)
			return
		}

		f, _ := os.Create("沪股列表.csv")
		wr := csv.NewWriter(f)
		wr.Comma = '\t'
		for _, record := range records {
			for i, s := range record {
				record[i] = strings.TrimSpace(s)

			}
			wr.Write(record)
		}
		f.Close()
	}

	// 从股票文件列表读出股票列表，写入数据库
	db, _ := sql.Open("mysql", "stockadmin:1973admin@tcp(127.0.0.1:3306)/stockdb")
	statement := "insert into stocklist (stockcode, stockname, stockmarket) values(?,?,?)"
	{ // 将深股列表里面的股票写入数据库
		writeRowToDB := func(s []string) {
			if len(s) < 7 {
				return
			}
			stockcode, stockname := s[5], s[6]
			_, err := db.Exec(statement, stockcode, stockname, "sz")
			if err != nil {
				log.Fatalln(err)
			}
		}
		DealWithCSVFile("./深股列表.csv", writeRowToDB)
	}
	{ // 将沪股列表里面的股票写入数据库
		writeRowToDB2 := func(s []string) {
			if len(s) < 2 {
				return
			}
			stockcode, stockname := s[2], s[3]
			//fmt.Println(stockcode, stockname)
			_, err := db.Exec(statement, stockcode, stockname, "sh")
			if err != nil {
				log.Fatalln(err)
			}
		}
		DealWithCSVFile("./沪股列表.csv", writeRowToDB2)
	}
}

// charset only one value is legal, that is "gb2312", all other is as "utf-8"
// if headerfile is a valid filename, FetchPage will use its data as header.
// headerfile 格式如下，可以直接从chrome浏览器拷贝出来，存成另外一个文件
// Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8
// Accept-Encoding: gzip, deflate
// ...
func FetchPage(url string, charset string, headerfile string) ([]byte, error) {

	// read headerfile into headers
	var headers [][]string
	if f, err := os.Open(headerfile); err == nil {
		rd := csv.NewReader(f)
		rd.FieldsPerRecord = -1
		rd.Comma = ':'
		headers, err = rd.ReadAll()
		if err != nil {
			fmt.Errorf("%v", err)
			headers = nil
		}

		f.Close()
	}

	// 创建http.Request结构对象并添加头信息
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Cant create Request obj: ", err)
		return nil, err
	}
	for _, h := range headers {
		key := h[0]
		val := h[1]
		if ok := strings.Contains(h[1], "http") || strings.Contains(h[1], "https"); ok {
			for i := 2; i < len(h); i++ {
				val += h[i]
			}
		}
		req.Header.Add(key, val)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Req err!:", err)
		return nil, err
	}
	if resp.Status[0] != '2' {
		return nil, errors.New(resp.Status)
	}
	defer resp.Body.Close()
	var reader io.Reader
	reader = resp.Body
	if charset == "gb2312" { // convert gb2312 into utf8
		reader = transform.NewReader(resp.Body, simplifiedchinese.GBK.NewDecoder())
	}
	buf := &bytes.Buffer{}
	_, err = io.Copy(buf, reader)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func QueryHTML(r io.Reader) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		log.Fatalln(err)
	}
	// find <div class="richContent" ...>
	//        <table ...>
	//          <tbody>
	// 内所有tr元素
	if trs := doc.Find(".richContent table tbody tr"); trs != nil {
		trs.Each(func(i int, n *goquery.Selection) {
			n.Find("td").Each(func(j int, c *goquery.Selection) {
				txt, _ := c.Find("span").Html()
				fmt.Printf("%s\t", txt)
			})
			fmt.Println()
		})
	}
}

type outputer func(s string)

func generateCSVFromXLSXBytes(bs []byte, sheetIndex int, csvDelimiter string, outputf outputer) error {

	xlFile, error := xlsx.OpenBinary(bs)
	if error != nil {
		return error
	}

	sheetLen := len(xlFile.Sheets)
	switch {
	case sheetLen == 0:
		return errors.New("This XLSX file contains no sheets.")
	case sheetIndex >= sheetLen:
		return fmt.Errorf("No sheet %d available, please select a sheet between 0 and %d\n", sheetIndex, sheetLen-1)
	}

	sheet := xlFile.Sheets[sheetIndex]

	for _, row := range sheet.Rows {
		var vals []string
		if row != nil {
			for _, cell := range row.Cells {
				//			str, err := cell.FormattedValue()
				// if err != nil {
				// 	vals = append(vals, err.Error())
				// }
				str := cell.Value
				vals = append(vals, fmt.Sprintf("%s", str))
			}
			outputf(strings.Join(vals, csvDelimiter) + "\n")
		}
	}

	return nil
}

type fnProcessCSVRecord func(csvRecord []string)

func DealWithCSVFile(name string, processRecord fnProcessCSVRecord) {
	f, err := os.Open(name)
	if err != nil {
		log.Fatal(err)
	}
	// rd := csv.NewReader(f)
	// rd.Comma = '\t'
	// rd.Read()
	// for {
	// 	cols, err := rd.Read()
	// 	if err != nil {
	// 		break
	// 	}
	// 	doRow(cols)
	// }
	rd := bufio.NewReader(f)
	rd.ReadString('\n')
	for {
		line, err := rd.ReadString('\n')
		if err != nil {
			break
		}
		line = line[:len(line)-1] //discard '\n'
		record := strings.Split(line, "\t")
		processRecord(record)
	}

	f.Close()
}
