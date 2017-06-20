package json

import (
	"errors"
	"fmt"
	"strconv"
)

type stateFn func(l *dictgen) stateFn

const (
	SPE_NONE = 0x00
	SPE_K    = 0x01
	SPE_C    = 0x02
	SPE_S    = 0x04
)

type Item struct {
	id    string
	t     string
	scope string
	desc  string
	v     string
	ri    int
}

func NewItem(id, t, scope, desc, v string, ri int) Item {
	return Item{id, t, scope, desc, v, ri}
}

type dictgen struct {
	ch     chan []Item
	output string

	done chan string
	err  chan error

	ri, ci int
}

func (l *dictgen) String() string {
	return l.output
}

func (l *dictgen) Push(items []Item) {
	l.ch <- items
}

func (l *dictgen) EndPush() {
	l.Push(make([]Item, 0))
}

func (l *dictgen) Done() string {
	return <- l.done
}

func (l *dictgen) Error() chan error {
	return l.err
}

func (l *dictgen) reset() {
	l.output = ""
	l.ri = 0
	l.ci = 0
}

func (l *dictgen) abort() {
	l.reset()

	l.done <- ""
}

func (l *dictgen) addItem(item *Item) error {
	if item.id == "" || item.v == "" {
		return nil
	}

	key := "\"" + item.id + "\""
	val := item.v

	if item.t == "int" {
		if _, err := strconv.Atoi(val); err != nil {
			return errors.New(fmt.Sprintf("unexpected type error, row:[%v] identi:%v define:int val:%v, but not a number. err : %v", item.ri+1, key, item.v, err))
		}

	} else if item.t == "decimal" {
		if _, err := strconv.ParseFloat(val, 32); err != nil {
			return errors.New(fmt.Sprintf("unexpected type error, row:[%v] identi:%v define decimal, but not a number. err : %v", item.ri, key, err))
		}

	} else {
		val = "\"" + val + "\""
	}

	object := key + ":" + val

	if l.ci > 0 {
		l.output += ","
		l.output += "\r\n"
	}

	l.output += "\t\t"

	l.output += object

	l.ci++

	return nil
}

func genNull(l *dictgen) stateFn {
	return nil
}

func genStart(l *dictgen) stateFn {
	l.output += "["
	l.output += "\r\n"

	return genNewRow
}

func genNewRow(l *dictgen) stateFn {
	row := <-l.ch

	if len(row) <= 0 {
		return genEnd
	}

	if l.ri > 0 {
		l.output += ","
		l.output += "\r\n"
	}

	l.output += "\t"

	l.output += "{"
	l.output += "\r\n"

	for _, item := range row {
		if err := l.addItem(&item); err != nil {
			l.err <- err
			fmt.Printf("%v\n", err)
			l.abort()

			return genNull
		}
	}

	l.output += "\r\n"
	l.output += "\t"
	l.output += "}"

	l.ci = 0

	l.ri++

	return genNewRow
}

func genEnd(l *dictgen) stateFn {
	l.output += "\r\n"
	l.output += "]"

	l.done <- l.output

	return nil
}

func NewGenerator() *dictgen {
	g := &dictgen{
		ch:   make(chan []Item, 1),
		done: make(chan string, 1),
		err:  make(chan error, 1),
	}

	go func(g *dictgen) {
		for state := genStart; state != nil; {
			state = state(g)
		}

		close(g.ch)
		close(g.done)
	}(g)

	return g
}
