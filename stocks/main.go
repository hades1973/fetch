// go build fetchAll.go
package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"sync"
)

const (
	DataDir           = "./"
	StockListFileName = "stocklist.csv"
)

func main() {
	if len(os.Args) == 2 {
		fmt.Println("usage: fetchstock")
		return
	}

	// encoding/csv 解码输入文件
	f, err := os.Open(path.Join(DataDir, StockListFileName))
	if err != nil {
		log.Fatal(err)
	}
	defer f.Sync()
	defer f.Close()
	reader := csv.NewReader(f)
	reader.Comma = ','
	reader.FieldsPerRecord = 2
	reader.TrimLeadingSpace = true

	// 读取股票列表文件，并行爬取数据
	done := make(chan struct{})
	var wg sync.WaitGroup // 记录活动的go routines
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println(err)
			return
		}
		wg.Add(1)
		go func(stockCode, stockName string) {
			defer wg.Done()
			UpdateStockData(stockCode, stockName)
			done <- struct{}{}
		}(record[0], record[1])

	}
	// 监视 go routines 返回, 然后关闭 done channel
	go func() {
		wg.Wait()
		close(done)
	}()
	for range done {
	}
}
