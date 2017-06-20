package json

import (
	"fmt"
	_ "strconv"

	"github.com/tealeg/xlsx"
)

type scaner struct {
	svrch chan string
	clich chan string
}

func (s *scaner) GetOutput() (string, string) {
	return <-s.svrch, <-s.clich
}

func (s *scaner) Close() {
	close(s.svrch)
	close(s.clich)
}

func NewScaner(xlfile *xlsx.File, dict_name string) *scaner {
	s := &scaner{
		svrch:   make(chan string, 1),
		clich:   make(chan string, 1),
	}

	go func(xlfile *xlsx.File) {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("recover in scaner ", r)
			}
		}()

		// only deal with Sheet 0
		sh := xlfile.Sheets[0]

		defer s.Close()

		if sh == nil {
			fmt.Println("cant find sheet, must something wrong happen.")
			return
		}

		nrows := len(sh.Rows)
		if nrows < 4 {
			fmt.Println("wrong fmt, 最少4行定义.")
			return
		}

		rows := sh.Rows

		// 前4行格式校验
		type_row := rows[1]
		scope_row := rows[2]

		// 类型校验
		for ci, cell := range type_row.Cells {
			text, _ := cell.String()

			if text != "int" && text != "string" && text != "decimal" {
				id, _ := sh.Cell(0, ci).String()
				fmt.Printf("invaild type definition, identi:%v type:%v\n", id, text)
				return
			}
		}

		// 主键、客户端、服务器字段校验
		m := make(map[int]int)
		for ci, cell := range scope_row.Cells {
			text, _ := cell.String()
			id, _ := sh.Cell(0, ci).String()

			f := m[ci]
			for _, c := range text {
				if c != 'k' && c != 's' && c != 'c' {
					fmt.Printf("invaild scope definition, identi:%v scope:%v\n", id, text)
					return
				}

				if c == 'k' {
					f |= SPE_K
				}else if c == 'c' {
					f |= SPE_C
				}else if c == 's' {
					f |= SPE_S
				}

				m[ci] = f
			}
		}

		sgen := NewGenerator()
		cgen := NewGenerator()

		for ri, row := range rows {
			if ri < 4 {
				continue
			}

			// end := true
			// for _, cell := range row.Cells {
			// 	if str, _ := cell.String(); str != "" {
			// 		end = false
			// 		break
			// 	}
			// }

			// if end {
			// 	sgen.EndPush()
			// 	cgen.EndPush()

			// 	break
			// }

			sitems := make([]Item, 0)
			citems := make([]Item, 0)

			for ci, cell := range row.Cells {
				k, _ := sh.Cell(0, ci).String()
				if k == "" {
					continue
				}

				t, _ := sh.Cell(1, ci).String()
				s, _ := sh.Cell(2, ci).String()
				d, _ := sh.Cell(3, ci).String()
				v, _ := cell.String()

				f := m[ci]

				item := NewItem(k, t, s, d, v, ri)

				if f & SPE_S != 0 {
					sitems = append(sitems, item)
				}

				if f & SPE_C != 0 {
					citems = append(citems, item)
				}
			}

			sgen.Push(sitems)
			cgen.Push(citems)
		}

		sgen.EndPush()
		cgen.EndPush()

		s.svrch <- sgen.Done()
		s.clich <- cgen.Done()

	}(xlfile)

	return s
}
