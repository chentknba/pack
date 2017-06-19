package json

import (
	"errors"
	"fmt"
	"strconv"
)

type stateFn func(l *generator) stateFn

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
	f     int
}

func NewItem(id, t, scope, desc, v string, ri, f int) Item {
	return Item{id, t, scope, desc, v, ri, f}
}

type generator struct {
	ch     chan []Item
	output string

	svout  string
	ciout  string
	svmeta string
	cimeta string

	svki int
	ciki int

	done chan struct{}
	err  chan error

	ri, ci int
}

func (l *generator) String() string {
	return l.output
}

func (l *generator) Push(items []Item) {
	l.ch <- items
}

func (l *generator) EndPush() {
	l.Push(make([]Item, 0))
}

func (l *generator) Done() chan struct{} {
	return l.done
}

func (l *generator) Error() chan error {
	return l.err
}

func (l *generator) reset() {
	l.output = ""
	l.ri = 0
	l.ci = 0
}

func (l *generator) abort() {
	l.reset()

	l.done <- struct{}{}
}

func (l *generator) addItem(item *Item) error {
	if item.id == "" || item.v == "" {
		return nil
	}

	key := "\"" + item.id + "\""
	val := item.v

	if item.t == "int" {
		if _, err := strconv.ParseInt(val, 10, 64); err != nil {
			return errors.New(fmt.Sprintf("unexpected type error, row:[%v] identi:%v define int, but not a number. err : %v", item.ri, key, err))
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

	if item.f&SPE_C > 0 {
	}

	if item.f&SPE_S > 0 {
	}

	l.output += object

	l.ci++

	return nil
}

func genNull(l *generator) stateFn {
	return nil
}

func genStart(l *generator) stateFn {
	l.output += "["
	l.output += "\r\n"

	return genNewRow
}

func genNewRow(l *generator) stateFn {
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

func genEnd(l *generator) stateFn {
	l.output += "\r\n"
	l.output += "]"

	l.svout = l.output
	l.ciout = l.output

	l.done <- struct{}{}

	return nil
}

func NewGenerator() *generator {
	g := &generator{
		ch:   make(chan []Item, 1),
		done: make(chan struct{}, 1),
		err:  make(chan error, 1),
	}

	go func(g *generator) {
		for state := genStart; state != nil; {
			state = state(g)
		}

		close(g.ch)
		close(g.done)
	}(g)

	return g
}
