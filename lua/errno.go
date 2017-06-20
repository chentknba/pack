package lua

import (
	"fmt"
	// "strconv"

	"github.com/tealeg/xlsx"
)

type errnogen struct {
	done chan string
}

func (g *errnogen) Done() string {
	return <-g.done
}

func NewErrnogen(xlfile *xlsx.File, lua_name string) *errnogen {
	g := &errnogen{
		done : make(chan string, 1),
	}

	go func(xlfile * xlsx.File) {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("recover in metagen ", r)
			}
		}()

		var errno string

		g.done <- errno

	}(xlfile)

	return g
}