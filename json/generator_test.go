package json_test

import (
    "testing"
    tojson "github.com/chentknba/pack/json"
)

func TestLex(t *testing.T) {
    lex := tojson.NewGenerator()

    go func() {
        items := make([]tojson.Item, 2)

        items = append(items, tojson.NewItem("id1", "int", "s", "desc", "100"))
        items = append(items, tojson.NewItem("col2", "string", "s", "desc", "xxx"))

        lex.Push(items)

        lex.Push(items)

        lex.Push(make([]tojson.Item, 0))
    }()

    lex.Wait()

    t.Error(lex)
}
