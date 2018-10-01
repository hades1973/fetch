package main

import (
	"database/sql"
	"encoding/csv"
	"log"
	"os"
	"path"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
)

type DayTick struct {
	stockcode string
	date      string
	open      float64
	high      float64
	low       float64
	close     float64
}

var db *sql.DB

func main() {
	db, _ = sql.Open("mysql", "stockadmin:1973admin@tcp(127.0.0.1:3306)/stockdb")
	statement := "select stockcode from stocklist"
	var daytick DayTick
	rows, err := db.Query(statement)
	if err != nil {
		log.Fatalln(err)
	}
	for rows.Next() {
		rows.Scan(&daytick.stockcode)
		f, err := os.Open(path.Join("~/diskD/", daytick.stockcode))
		if err != nil {
			continue
		}
		rd := csv.NewReader(f)
		record, err := rd.Read() // skip comment
		record, err = rd.Read()
		for err == nil {
			daytick.date = record[0]
			daytick.open, _ = strconv.ParseFloat(record[1], 64)
			daytick.high, _ = strconv.ParseFloat(record[2], 64)
			daytick.low, _ = strconv.ParseFloat(record[3], 64)
			daytick.close, _ = strconv.ParseFloat(record[4], 64)
			db.Exec("insert into dayticks (stockcode, date, open, high, low, close) values(?,?,?,?,?,?)",
				daytick.stockcode,
				daytick.date,
				daytick.open,
				daytick.high,
				daytick.low,
				daytick.close,
			)
			record, err = rd.Read()
		}

		f.Close()
	}
	rows.Close()

	return
}
