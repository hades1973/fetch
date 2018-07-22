package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	iconv "gopkg.in/iconv.v1"
)

var location = time.Now().Location()
var logf = os.Stdin

var limitUpdateStockDatas = make(chan struct{}, 5)

func UpdateStockData(stockcode, stockname string) {
	limitUpdateStockDatas <- struct{}{}
	defer func() {
		<-limitUpdateStockDatas
	}()

	// f for 股票数据文件，lastUpdate for 最后更新日期
	var (
		f          *os.File
		lastUpdate string
	)

	// 如果log文件不存在，则直接将2004-01-01作为最后更新日期，并创建股票数据文件f
	// 否则从log文件读取最后的更新日期，并打开已经存在的股票文件
	bytes, err := ioutil.ReadFile(path.Join(DataDir, stockcode+".log"))
	if err != nil {
		lastUpdate = fmt.Sprintf("%s", "2004-01-01")
		f, err = os.Create(path.Join(DataDir, stockcode+".csv"))
		if err != nil {
			fmt.Fprintf(logf, "%s, %s\n", stockcode, err)
			return
		}
		fmt.Fprintf(f, "date,open,high,close,low,volumn,transaction,pow\n")
	} else {
		lastUpdate = strings.Fields(string(bytes))[0]
		f, err = os.OpenFile(path.Join(DataDir, stockcode+".csv"), os.O_APPEND|os.O_WRONLY, 0666)
		if err != nil {
			log.Fatal(err)
		}
	}
	fmt.Printf("Update stock \"%s\" data\n", stockname)
	defer f.Close()

	// 从网页爬取数据
	year, month, _ := time.Now().Date()
	season := MonthToSeason(month)
	lines := make([]string, 0)
	var finished bool
FETCHUPDATE:
	for year >= 2004 {
		for season >= 1 {
			lines, finished = fetchUpdate(lines, stockcode, year, season, lastUpdate)
			if finished {
				break FETCHUPDATE
			}

			season--
		}
		season = 4
		year--
	}
	if len(lines) == 0 { // 没有新的更新直接返回
		return
	}

	// 从网页爬回的数据是从后先前排列期，将其逆序写入文件
	// 需要对价格除权
	var date string
	var deal [7]float64 // for open,high,close,low,volumn,transaction,pow
	for i := len(lines) - 1; i >= 0; i-- {
		rec := strings.FieldsFunc(lines[i], func(r rune) bool {
			if r == ',' {
				return true
			}
			return false
		})
		date = rec[0]
		rec = rec[1:]
		for i, v := range rec {
			fmt.Sscan(v, "%f", deal[i])
		}
		for i := 0; i < 4; i++ {
			deal[i] /= deal[6]
		}
		fmt.Fprintf(f, "%s,%f,%f,%f,%f,%f,%f,%f\r\n",
			date, deal[0], deal[1], deal[2], deal[3], deal[4], deal[5], deal[6])
	}

	// 将最后交易日期写入log文件，以便更新
	flog, err := os.Create(path.Join(DataDir, stockcode+".log"))
	if err != nil {
		fmt.Println("Can't create log file for %s\n", stockcode)
	}
	defer flog.Close()
	fmt.Fprintf(flog, "%s\n", lines[0])

	fmt.Printf("Finished update stock \"%s\"\n", stockname)
}

func MonthToSeason(month time.Month) (season int) {
	if month == 1 || month == 2 || month == 3 {
		season = 1
	} else if month == 4 || month == 5 || month == 6 {
		season = 2
	} else if month == 7 || month == 8 || month == 9 {
		season = 3
	} else {
		season = 4
	}
	return season
}

func fetchData(lines []string, stockcode string, year, season int) []string {
	htmlPage, err := fetchStockHTMLPageFromSina(stockcode, year, season)
	if err != nil || len(htmlPage) == 0 {
		return lines
	}
	table, ok := GetStockTable(htmlPage, "FundHoldSharesTable")
	if ok != true || len(table) == 0 {
		return lines
	}
	rowscols := ParseRowsColsFromTable(table)
	for _, row := range rowscols {
		line := strings.Join(row, ",")
		ch, _ := utf8.DecodeRune([]byte(line))
		if unicode.IsDigit(ch) != true {
			continue
		}
		lines = append(lines, line)
	}
	return lines
}

func fetchUpdate(lines []string, stockcode string, year, season int, lastUpdate string) ([]string, bool) {
	htmlPage, err := fetchStockHTMLPageFromSina(stockcode, year, season)
	if err != nil || htmlPage == nil {
		return lines, false
	}

	table, ok := GetStockTable(htmlPage, "FundHoldSharesTable")
	if ok != true || len(table) == 0 {
		return lines, false
	}

	rowscols := ParseRowsColsFromTable(table)
	for _, colsOfRow := range rowscols {
		line := strings.Join(colsOfRow, ",")
		ch, _ := utf8.DecodeRune([]byte(line))
		if unicode.IsDigit(ch) != true {
			continue
		}
		if colsOfRow[0] <= lastUpdate {
			fmt.Printf("%s\n%s\n", colsOfRow, lastUpdate)
			return lines, true
		}
		lines = append(lines, line)
	}
	return lines, false
}

func fetchStockHTMLPageFromSina(stockcode string, year, season int) ([]byte, error) {
	const FMTURL = "http://vip.stock.finance.sina.com.cn/corp/go.php/vMS_FuQuanMarketHistory/stockid/%s.phtml?year=%d&jidu=%d"
	url := fmt.Sprintf(FMTURL, stockcode, year, season)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Print("fetch page failed")
		return nil, err
	}
	defer resp.Body.Close()

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
