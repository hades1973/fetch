package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path"

	_ "github.com/go-sql-driver/mysql"
)

//type DayTick struct {
//	stockcode string
//	date      string
//	open      float64
//	high      float64
//	low       float64
//	close     float64
//	//	average   float64 // 均价=成交额/成交量
//	volumn float64
//	pow    float64
//}

var db *sql.DB

func main() {
	// 检查原始数据目录是否存在
	//if fileInfo, err := os.Stat("stockdata"); err != nil {
	//	log.Fatalln(err)
	//} else {
	//	if fileInfo.IsDir() != true {
	//		log.Fatalln("No find data path")
	//	}
	//}

	var err error
	db, err = sql.Open("mysql", "stocker:1973stocker@tcp(127.0.0.1:3306)/stockdb")
	if err != nil {
		log.Fatalln(err)
	}
	rows, err := db.Query(`select stockcode from stocklist`)
	if err != nil {
		log.Fatalln(err)
	}

	var (
		cnt       int = -1
		tx        *sql.Tx
		stmt      *sql.Stmt
		stockcode string
	)

BEGIN_INSERT:
	for {
		if ok := rows.Next(); ok != true {
			//			fmt.Printf("%v\n", rows.Err())
			fmt.Println("Leave out loop")
			fmt.Println(cnt)
			if ((cnt % 1000000) > 0) && ((cnt % 100000) < 99999) {
				tx.Commit()
			}
		}
		err := rows.Scan(&stockcode)
		if err != nil {
			log.Fatalln(err)
		}
		f, err := os.Open(path.Join("./stockdata/", stockcode+".csv"))
		if err != nil {
			f.Close()
			continue BEGIN_INSERT
		}
		rd := csv.NewReader(f)
		rd.Comment = '#'
		rd.Comma = ','
		for {
			recs, err := rd.Read()
			if err != nil {
				break
			}
			cnt++
			fmt.Println(cnt % 100000)
			if (cnt % 100000) == 0 {
				fmt.Println("db.begin() and prepare")
				tx, _ = db.Begin()
				stmt, err = tx.Prepare("insert into stockdayticks (stockcode, thedate, theopen, thehigh, thelow, theclose, thevolumn, thepow) values(?,?,?,?,?,?, ?, ?)")
				if err != nil {
					log.Fatalln(err)
				}
			} else if (cnt % 100000) == 99999 {
				fmt.Println("db.commit()")
				tx.Commit()
			}

			li := make([]string, 0)
			li = append(li, "'"+stockcode+"'")
			for i, rec := range recs {
				recs[i] = "'" + rec + "'"
				li = append(li, recs[i])
			}
			_, err = stmt.Exec(li[0], li[1], li[2], li[3], li[4], li[5], li[6], li[7])
			if err != nil {
				fmt.Println(err)
			}
		}

		f.Close()
	}
	rows.Close()

	db.Close()
	return
}
