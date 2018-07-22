// 从凤凰财经网站抓取股票数据，该数据没有权值。
package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

const (
	DataDir = "/home/jns/diskD/stockdata/"
)

var StockListFileName = path.Join(DataDir, "stocklist.csv")

// main func //////////////////////////////////////////////
func main() {
	// 从股票列表文件读取所有股票的代码、名称、最后更新日期、最后权值
	// all is a list of []string, its element is a list:[stockcode, stockname, lastUpdateDay, lastPower]
	all := GetStockList()
	if all == nil {
		fmt.Println(nil)
		return
	}
	// 逐个股票更新数据
	var (
		i    int
		done = make(chan struct{})
	)
	for i, _ = range all {
		go func(stock []string) {
			UpdateStockData(stock)
			done <- struct{}{}
		}(all[i])
	}

	for ; i > 0; i-- {
		<-done
	}

	UpdateStockList(all)
}

// GetStockList func //////////////////////////////////////////
// 获取所有股票代码、名称、最后更新日期、最后一次更新时的权值
func GetStockList() [][]string {
	var (
		f   *os.File
		err error
	)
	f, err = os.Open(StockListFileName)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	rd := csv.NewReader(f)
	rd.Comma = ','
	rd.TrimLeadingSpace = true
	allList, err := rd.ReadAll()
	if err != nil {
		log.Fatal(err)
	}
	return allList
}

// UpdateStockList func ///////////////////////////////////////
// 根据更新后的股票数据更新stocklist.csv文件
func UpdateStockList(all [][]string) {
	var (
		f   *os.File
		err error
	)
	f, err = os.Create(StockListFileName)
	if err != nil { // 如果出错会丢失原有的股票数据文件内容，要常备份!!!!
		log.Fatal(err)
	}
	defer f.Close()
	wr := csv.NewWriter(f)
	wr.Comma = ','
	for _, item := range all {
		wr.Write(item)
	}
	wr.Flush()
	return
}

// UpdateStockData func ///////////////////////////////////////
type DealJson struct {
	Records [][]string `json:"record"`
}

var limitedRoutines = make(chan struct{}, 3)

func UpdateStockData(stock []string) { // [code, name, date, pow] for stock
	code, name, lastDate, lastPow := stock[0], stock[1], stock[2], stock[3]
	var stockFileName = path.Join(DataDir, code+".csv")

	// 限制并行数量，以防止被网站封堵
	limitedRoutines <- struct{}{}
	defer func() {
		<-limitedRoutines
	}()

	// 根据股票代码构造申请网页的url
	var (
		urlFmt string = "http://api.finance.ifeng.com/akdaily/?code=%s%s&type=last"
		url    string
		page   []byte
		err    error
	)
	if code[0] == '0' {
		url = fmt.Sprintf(urlFmt, "sz", code)
	} else if code[0] == '6' {
		url = fmt.Sprintf(urlFmt, "sh", code)
	} else {
		fmt.Println(code, " is not in sh or sz")
		return
	}
	// 从凤凰财经爬起原始的json字串数据
	page = fetchPageFromIfeng(url)
	if page == nil {
		return
	}
	// 解析json字串
	var dealJson DealJson
	err = json.Unmarshal(page, &dealJson)
	if err != nil {
		fmt.Println(err)
		return
	}
	//去除数字内部的逗号
	deal := dealJson.Records
	for i := 0; i < len(deal); i++ {
		for j := 0; j < len(deal[i]); j++ {
			strings.Replace(deal[i][j], ",", "", -1)
		}
	}
	// find data after lastDate, those will write into stock file
	var i int = 0
	for ; i < len(deal); i++ {
		if lastDate == deal[i][0] {
			i++
			break
		}
	}
	if i == len(deal) { // 最后一次更新时间与爬取的数据的最后时间一致，无需更新数据
		return
	}
	deal = deal[i:] // 需要更新的数据
	// 将需要更新的数据写入股票数据文件
	var fstock *os.File
	fstock, err = os.OpenFile(stockFileName, os.O_WRONLY|os.O_APPEND, 0777)
	if err != nil {
		return
	}
	writer := csv.NewWriter(fstock)
	for _, words := range deal {
		// 只需要：date,open,high,close,low,volumn,power，其中power来自lastPow
		err = writer.Write([]string{
			words[0], words[1], words[2], words[3], words[4], words[5], lastPow})
		if err != nil {
			fmt.Println(err)
		}
	}
	writer.Flush()
	fstock.Close()

	// 更新股票记录的最后更改日期
	stock[2] = deal[len(deal)-1][0]

	// 打印更新完毕信息，并休眠10秒以免被网站封杀
	fmt.Println("updated ", name)
	time.Sleep(10 * time.Second)
	return
}

// fetchPageFromIfeng func ////////////////////////////////////////////////////
// 从凤凰财经网站获取股票原始网页内容，是个json格式的字串
func fetchPageFromIfeng(url string) []byte { // 返回nil表示抓取网页失败
	var (
		b []byte
		i int
	)
	for i = 3; i > 0; i-- { // 顶多抓取网页三次
		resp, err := http.Get(url)
		if err != nil {
			resp.Body.Close()
			time.Sleep(2 * time.Second)
			continue
		}
		if b, _ = ioutil.ReadAll(resp.Body); strings.Index(string(b), "2") < 0 {
			b = nil
			resp.Body.Close()
			time.Sleep(2 * time.Second)
			continue
		}
		resp.Body.Close()
		break
	}
	return b
}
