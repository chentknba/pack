package lua

import (
	"fmt"
	"strconv"

	"github.com/tealeg/xlsx"
)

type metagen struct {
	done chan string
}

func (g *metagen) Done() string {
	return <-g.done
}

func NewMetagen(xlfile *xlsx.File, dict_name string) *metagen {
	g := &metagen{
		done : make(chan string, 1),
	}

	go func(xlfile * xlsx.File) {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("recover in metagen ", r)
			}
		}()

		sh := xlfile.Sheets[0]

		// defer s.Close()

		if sh == nil {
			fmt.Println("cant find sheet, must something wrong happen.")
			return
		}

		nrows := len(sh.Rows)
		if nrows < 4 {
			fmt.Println("wrong fmt, 最少4行定义.")
			return
		}

		var meta string
		pathstr := "path = " + "\"" + dict_name + "\""
		fmtstr  := "format = \"json\""

		meta += "\t"
		meta += dict_name
		meta += " = "
		meta += "\r\n"
		meta += "\t"
		meta += "{"
		meta += "\r\n"
		meta += "\t\t"
		meta += pathstr
		meta += ","
		meta += "\r\n"
		meta += "\t\t"
		meta += fmtstr
		meta += ","

		keynum := 1

		scope_row := sh.Rows[2]
		for ci, cell := range scope_row.Cells {
			text, _ := cell.String()
			id, _ := sh.Cell(0, ci).String()

			iskey := false
			for _, c := range text {
				if  c == 'k' {
					iskey = true
					break
				}
			}

			if iskey {
				keyi := "key" + strconv.Itoa(keynum)
				keystr := keyi + " = " + "\"" + id + "\""

				meta += "\r\n"
				meta += "\t\t"
				meta += keystr
				meta += ","

				keynum++
			}
		}

		meta += "\r\n"
		meta += "\t"
		meta += "}"

		g.done <- meta

	}(xlfile)

	return g
}

