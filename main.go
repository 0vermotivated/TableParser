package main

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"os"
	"strings"

	"golang.org/x/net/html"
)

const URL string = "https://ru.wikipedia.org/wiki/%D0%A2%D0%B0%D0%B1%D0%BB%D0%B8%D1%86%D0%B0"

func findTable(n *html.Node) *html.Node {
	if n.Type == html.ElementNode && n.Data == "table" {
		return n
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if result := findTable(c); result != nil {
			return result
		}
	}
	return nil
}

func parseTable(table *html.Node) [][]string {
	var rows [][]string
	for n := table.FirstChild; n != nil; n = n.NextSibling {
		if n.Type == html.ElementNode && (n.Data == "thead" || n.Data == "tbody" || n.Data == "tr") {
			rows = append(rows, parseRows(n)...)
		}
	}
	return rows
}

func parseRows(node *html.Node) [][]string {
	var rows [][]string
	if node.Type == html.ElementNode && node.Data == "tr" {
		var row []string
		for td := node.FirstChild; td != nil; td = td.NextSibling {
			if td.Type == html.ElementNode && (td.Data == "td" || td.Data == "th") {
				cellText := getText(td)
				row = append(row, cellText)
			}
		}
		if len(row) > 0 {
			rows = append(rows, row)
		}
	} else {
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			rows = append(rows, parseRows(c)...)
		}
	}
	return rows
}

func getText(n *html.Node) string {
	if n.Type == html.TextNode {
		return strings.TrimSpace(n.Data)
	}
	var result string
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		result += getText(c)
	}
	return strings.TrimSpace(result)
}

func main() {
	url := URL
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Ошибка загрузки: %v\n", err)
		return
	}
	defer resp.Body.Close()

	doc, err := html.Parse(resp.Body)
	if err != nil {
		fmt.Printf("Ошибка на разборе хтмл: %v\n", err)
		return
	}

	table := findTable(doc)
	if table == nil {
		fmt.Println("Тег таблицы не найден.")
		return
	}

	htmlFile, err := os.Create("table.html")
	if err != nil {
		fmt.Printf("Ошибка создания хтмл-фалйа: %v\n", err)
		return
	}
	defer htmlFile.Close()
	html.Render(htmlFile, table)
	fmt.Println("Таблица сохранена")

	rows := parseTable(table)
	if len(rows) == 0 {
		fmt.Println("Данные таблицы не найдены.")
		return
	}

	csvFile, err := os.Create("table.csv")
	if err != nil {
		fmt.Printf("Ошибка создания CSV-файла: %v\n", err)
		return
	}
	defer csvFile.Close()

	if _, err := csvFile.Write([]byte{0xEF, 0xBB, 0xBF}); err != nil {
		fmt.Printf("Ошибка в CSV: %v\n", err)
		return
	}

	writer := csv.NewWriter(csvFile)
	defer writer.Flush()

	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			fmt.Printf("Ошибка записи в CSV: %v\n", err)
		}
	}
	fmt.Println("Таблица сохранена в csv")
}
