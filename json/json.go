package json

type stateFn func(l *lexer) stateFn

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

type lexer struct {
	ch     chan []Item
	output string

	svout string
	ciout string

	done chan struct{}
	err  error

	ri, ci int
}

func (l *lexer) String() string {
	return l.output
}

func (l *lexer) Push(items []Item) {
	l.ch <- items
}

func (l *lexer) Wait() {
	<-l.done
}

func (l *lexer) reset() {
	l.output = ""
	l.ri = 0
	l.ci = 0
}

func (l *lexer) abort() {
	l.reset()

	l.done <- struct{}{}
}

func (l *lexer) addItem(item *Item) error {
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

func nullState(l *lexer) stateFn {
	return nil
}

func startState(l *lexer) stateFn {
	l.output += "["

	return newRowState
}

func newRowState(l *lexer) stateFn {
	row := <-l.ch

	if len(row) <= 0 {
		return endState
	}

	if l.ri > 0 {
		l.output += ","
	}

	l.output += "{"

	for _, item := range row {
		if err := l.addItem(&item); err != nil {
			l.err = err
			l.abort()

			return nullState
		}
	}

	l.output += "}"
	l.ci = 0

	l.ri++

	return newRowState
}

func endState(l *lexer) stateFn {
	l.output += "]"

	l.done <- struct{}{}

	return nil
}

func NewLexer() *lexer {
	l := &lexer{
		ch:   make(chan []Item),
		done: make(chan struct{}),
	}

	s := startState

	go func(l *lexer) {
		for {
			if s = s(l); s == nil {
				break
			}
		}
	}(l)

	return l
}
