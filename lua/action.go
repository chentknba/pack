package lua

import (
	_ "fmt"

	"github.com/tealeg/xlsx"
)

type actiongen struct {
	m map[string]string
}

// Done
func (g *actiongen) Done() map[string]string {
	return g.m
}

// gentype 生成action_type
func gentype(sh *xlsx.Sheet) string {
	var action string

	action += AUTO_FILE_DESC
	action += "\r\n"
	action += "\r\n"
	action += "local action_type = "
	action += "\r\n"
	action += "{"
	action += "\r\n"

	m := make(map[string]int)
	for ci, cell := range sh.Rows[0].Cells {
		text, _ := cell.String()
		m[text] = ci
	}

	rows := sh.Rows

	for ri, _ := range rows {
		if ri < 3 {
			continue
		}

		var action_name, action_value, content, desc string

		action_name, _ = sh.Cell(ri, m["action_name"]).String()
		action_value, _ = sh.Cell(ri, m["action_value"]).String()
		content, _ = sh.Cell(ri, m["content"]).String()
		desc, _ = sh.Cell(ri, m["desc"]).String()

		action += TAB_1
		action += COMMENT
		action += " "
		action += content
		if desc != "" {
			action += "("
			action += desc
			action += ")"
		}

		action += "\r\n"
		action += TAB_1
		action += action_name
		action += TAB_10
		action += "="
		action += " "
		action += "\"" + action_value + "\""
		action += ","
		action += "\r\n"
	}

	action += "\r\n"
	action += "}"
	action += "\r\n"
	action += "\r\n"

	action += "return action_type"

	return action
}

func genevent(sh *xlsx.Sheet) string {
	var event string

	event += AUTO_FILE_DESC
	event += "\r\n"
	event += "\r\n"
	event += "local action_event = "
	event += "\r\n"
	event += "{"
	event += "\r\n"

	m := make(map[string]int)
	for ci, cell := range sh.Rows[0].Cells {
		text, _ := cell.String()
		m[text] = ci
	}

	rows := sh.Rows

	for ri, _ := range rows {
		if ri < 3 {
			continue
		}

		var event_name, event_prefix, event_suffix, content, desc string

		event_name, _ = sh.Cell(ri, m["event_name"]).String()
		event_prefix, _ = sh.Cell(ri, m["event_prefix"]).String()
		event_suffix, _ = sh.Cell(ri, m["event_suffix"]).String()
		content, _ = sh.Cell(ri, m["content"]).String()
		desc, _ = sh.Cell(ri, m["desc"]).String()

		event += TAB_1
		event += COMMENT
		event += " "
		event += content
		if desc != "" {
			event += "("
			event += desc
			event += ")"
		}

		event += "\r\n"
		event += TAB_1
		event += event_name
		event += TAB_10
		event += "="
		event += " "
		event += "\"" + event_prefix + event_suffix + "\""
		event += ","
		event += "\r\n"
	}

	event += "\r\n"
	event += "}"
	event += "\r\n"
	event += "\r\n"

	event += "return action_event"

	return event
}

func NewActiongen(xlfile *xlsx.File) *actiongen {
	g := &actiongen{
		m : make(map[string]string),
	}

	for _, sh := range xlfile.Sheets {
		name := sh.Name

		var content string

		if name == "action_type" {
			content = gentype(sh)
		}else if name == "action_event" {
			content = genevent(sh)
		}else{
			continue
		}

		g.m[name] = content
	}

	return g
}