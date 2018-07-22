package main

import (
	"fmt"
	"bytes"
	"strings"
)

// GetStockTable returns content between "<table id=id ...>content</table>
func GetStockTable(page []byte, id string) (contentInsdeTag []byte, ok bool) {
	begin := fmt.Sprintf("<table id=\"%s\">", id)
	end := `</table>`
	i := bytes.Index(page, []byte(begin))
	if i == -1 {
		return nil, false
	}
	page = page[i+len(begin):]
	i = bytes.Index(page, []byte(end))
	if i == -1 {
		return nil, false
	}
	return page[:i], true
}

func ParseRowsColsFromTable(table []byte) (rowscols [][]string) {
	rows := ParseRowsOrColsFromHTMLTable(table, "tr")
	if len(rows) == 0 {
		return nil
	}
	for i := 1; i < len(rows); i++ { // skip first row, which is comment
		row := rows[i]
		cols := ParseRowsOrColsFromHTMLTable([]byte(row), "td")
		if len(cols) == 0 {
			continue
		}
		for k, col := range cols {
			cols[k] = strings.TrimSpace(StripData(col))
		}
		if len(cols[0]) == 0 {
			continue
		}
		rowscols = append(rowscols, cols)
	}
	return rowscols
}

func ParseRowsOrColsFromHTMLTable(table []byte, whichTag string) (rowsORcols []string) {
	for {
		i := bytes.Index(table, []byte("<"+whichTag))
		if i == -1 {
			break
		}
		table = table[i+len("<"+whichTag):]
		i = bytes.Index(table, []byte(">"))
		if i == -1 {
			break
		}
		table = table[i+len(">"):]
		i = bytes.Index(table, []byte("</"+whichTag+">"))
		if i == -1 {
			break
		}

		rowsORcols = append(rowsORcols, string(table[:i]))
		table = table[i+len("</"+whichTag+">"):]
	}
	return rowsORcols
}

func StripData(html string) (data string) {
	begin := `<`
	escap := `</`
	close := `>`
	tag := ""
	for {
		i := strings.Index(html, begin)
		if i == -1 {
			break
		}
		html = html[i+len(begin):]
		j := strings.Index(html, " ")
		if j != -1 {
			tag = html[:j]
			html = html[j+len(" "):]
		} else {
			j = strings.Index(html, close)
			if j == -1 {
				break
			}
			tag = html[:j]
			html = html[j+len(close):]
		}
		j = strings.Index(html, close)
		if j == -1 {
			break
		}
		html = html[j+len(close):]
		k := strings.LastIndex(html, escap+tag+close)
		if k == -1 {
			break
		}
		html = html[:k]
	}
	return html
}
