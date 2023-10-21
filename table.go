package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"strconv"
)

var table_html = "table.html"
var table_template = "table.tmpl"
var funcs = template.FuncMap{"mod": func(i, j int) bool { return i%j == 0 }}
var templates = template.Must(template.New("table.html").Funcs(funcs).ParseFiles("./static/templates/"+table_template, "./static/templates/"+table_html))

type Checker struct {
	Id string `json:"id"`
	//Class string `json:"cl"`
}

type Td struct {
	Id    string  `json:"id"`
	Class string  `json:"cl"`
	Ch    Checker `json:"ch"`
}

type Tr struct {
	Th  int  `json:"th"`
	Tds []Td `json:"tds"`
}

type Table struct {
	Name      string     `json:"n"`
	Trs       []Tr       `json:"trs"`
	Whodo     string     `json:"whodo"`
	NextSteps []NextStep `json:"next"`
}

type NextStep struct {
	Checker   string   `json:"ch"`
	NextSteps []string `json:"steps"`
	NextKicks []string `json:"kicks"`
}

func renderTemplate(wr io.Writer, tmpl string, data any) {
	err := templates.ExecuteTemplate(wr, tmpl, data)
	if err != nil {
		panic(err)
	}
}

func renderJson(wr io.Writer, state any) {
	b, err := json.Marshal(state)
	if err != nil {
		panic(err)
	}
	wr.Write(b)
}

func initTable(isexist func(chid string) bool) *Table {
	if isexist == nil {
		isexist = func(chid string) bool { return false }
	}
	t := Table{
		Name:  "Checkers table",
		Whodo: "checker-white",
	}
	trs := make([]Tr, 8, 8)
	w := 1
	b := 1
	for i := 8; i >= 1; i-- {
		k := 1
		tr := Tr{
			Th: i,
		}
		tds := make([]Td, 8, 8)
		for j := 'a'; j <= 'h'; j++ {
			td := Td{
				// do not forget that "j" is not string because of literal
				// Id: fmt.Sprint(i) + string(j),
				Id: strconv.Itoa(i) + string(j),
			}
			if i%2+1 == k%2+1 || i%2 == k%2 {
				td.Class = "square black"

				if i >= 6 && i <= 8 {
					chid := "b" + fmt.Sprint(b)
					if isexist(chid) {
						td.Ch = Checker{
							Id: chid,
							//Class: "checker-white",
						}
					}
					b++
				} else if i >= 1 && i <= 3 {
					chid := "w" + fmt.Sprint(w)
					if isexist(chid) {
						td.Ch = Checker{
							Id: chid,
							//Class: "checker-black",
						}
					}
					w++
				}

			} else {
				td.Class = "white"
			}
			tds[k-1] = td
			k++

		}
		tr.Tds = tds
		trs[8-i] = tr
	}
	t.Trs = trs

	return &t
}
