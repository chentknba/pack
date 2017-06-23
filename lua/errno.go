package lua

import (
	_ "fmt"

	_ "sync"
	"github.com/tealeg/xlsx"
)

// var wg sync.WaitGroup

type errnogen struct {
	m map[string]string
}

func (g *errnogen) Done() map[string]string {
	return g.m
}

// func gen(sh *xlsx.Sheet, r chan result) {
func gen(sh *xlsx.Sheet) result {
	var errno string

	m := make(map[string]int)
	for ci, cell := range sh.Rows[0].Cells {
		text, _ := cell.String()
		m[text] = ci
	}

	rows := sh.Rows

	for ri, _ := range rows {
		if ri < 4 {
			continue
		}

		var id, name, content, params, desc string

		id, _ = sh.Cell(ri, m["id"]).String()
		name, _ = sh.Cell(ri, m["name"]).String()
		content, _ = sh.Cell(ri, m["content"]).String()
		params, _ = sh.Cell(ri, m["params"]).String()
		desc, _ = sh.Cell(ri, m["desc"]).String()

		errno += TAB_1
		errno += name
		errno += TAB_10
		errno += "="
		errno += " "
		errno += id
		errno += ","
		errno += TAB_2
		errno += COMMENT
		errno += " "
		errno += content
		errno += "("
		if params != "" {
			errno += params
		}

		if desc != "" {
			errno += desc
		}

		errno += ")"
	}

	return result{file:sh.Name, content:errno}
	// r <- result{file:sh.Name, content:errno}
}

func NewErrnogen(xlfile *xlsx.File) *errnogen {
	g := &errnogen{
		m : make(map[string]string),
	}

	// ch := make(chan result)
	// defer close(ch)

	for _, sh := range xlfile.Sheets {
		// wg.Add(1)

		// go gen(sh, ch)

		// result := <-ch

		result := gen(sh)

		g.m[result.file] = result.content
	}

	// wg.Wait()

	return g
}
