package json

type stateFn func(l *generator) stateFn

type Item struct {
	id    string
	t     string
	scope string
	desc  string
	v     string
}

func NewItem(id, t, scope, desc, v string) Item {
	return Item{id, t, scope, desc, v}
}

type generator struct {
	ch     chan []Item
	output string

	svout string
	ciout string

	done chan struct{}
	err  error

	ri, ci int
}

func (l *generator) String() string {
	return l.output
}

func (l *generator) Push(items []Item) {
	l.ch <- items
}

func (l *generator) Wait() {
	<-l.done
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

	if item.t == "int" || item.t == "decimal" {
	} else {
		val = "\"" + val + "\""
	}

	object := key + ":" + val

	if l.ci > 0 {
		l.output += ","
	}

	for c := range item.scope {
		if c == 's' {
		} else if c == 'c' {
		} else if c == 'k' {
		}
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

	return genNewRow
}

func genNewRow(l *generator) stateFn {
	row := <-l.ch

	if len(row) <= 0 {
		return genEnd
	}

	if l.ri > 0 {
		l.output += ","
	}

	l.output += "{"

	for _, item := range row {
		if err := l.addItem(&item); err != nil {
			l.err = err
			l.abort()

			return genNull
		}
	}

	l.output += "}"
	l.ci = 0

	l.ri++

	return genNewRow
}

func genEnd(l *generator) stateFn {
	l.output += "]"

	l.done <- struct{}{}

	return nil
}

func NewGenerator() *generator {
	g := &generator{
		ch:   make(chan []Item),
		done: make(chan struct{}),
	}

	go func(g *generator) {
		for state := genStart; state != nil; {
			state = state(g)
		}
	}(g)

	return g
}
