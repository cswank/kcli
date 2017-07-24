package views

import (
	"fmt"
	"time"
)

type helpStep struct {
	caption  string
	template string
	args     []interface{}
	duration time.Duration
}

var (
	jumpSteps []helpStep
)

func getJumpSteps() []helpStep {
	header := c1("partition     1st offset             current offset         last offset            size")
	rows := c2(`0             0                      %s                   1100                   1100
1             55                     %s                   2405                   2320`)
	tpl := fmt.Sprintf(`%s
%s









%%s`, header, rows)

	return []helpStep{
		{
			caption:  "when you jump on a topic the number you enter",
			template: tpl,
			args:     []interface{}{c2("0   "), c2("55  "), ""},
			duration: 2 * time.Second,
		},
		{
			caption:  "when you jump on a topic the number you enter",
			template: tpl,
			args:     []interface{}{c2("0   "), c2("55  "), c1("jump: ")},
			duration: 400 * time.Millisecond,
		},
		{
			caption:  "is relative to 1st offset",
			template: tpl,
			args:     []interface{}{c2("0   "), c2("55  "), c1("jump: 4")},
			duration: 400 * time.Millisecond,
		},
		{
			caption:  "is relative to 1st offset",
			template: tpl,
			args:     []interface{}{c2("0   "), c2("55  "), c1("jump: 40")},
			duration: 400 * time.Millisecond,
		},
		{
			caption:  "is relative to 1st offset",
			template: tpl,
			args:     []interface{}{fmt.Sprintf(c2("%s   "), c3("0")), fmt.Sprintf(c2("%s  "), c3("55")), c1("jump: 40")},
			duration: 1000 * time.Millisecond,
		},
		{
			caption:  "is relative to 1st offset",
			template: tpl,
			args:     []interface{}{fmt.Sprintf(c2("%s  "), c3("40")), fmt.Sprintf(c2("%s  "), c3("95")), c1("jump: 40")},
			duration: 2 * time.Second,
		},
		{
			caption:  "",
			template: tpl,
			args:     []interface{}{c2("40  "), c2("95  "), c1("")},
			duration: 1 * time.Second,
		},
		{
			caption:  "a negative number is relative the the last offset",
			template: tpl,
			args:     []interface{}{c2("40  "), c2("95  "), c1("jump:")},
			duration: 1 * time.Second,
		},
		{
			caption:  "a negative number is relative the the last offset",
			template: tpl,
			args:     []interface{}{c2("40  "), c2("95  "), c1("jump: -")},
			duration: 400 * time.Millisecond,
		},
		{
			caption:  "a negative number is relative the the last offset",
			template: tpl,
			args:     []interface{}{c2("40  "), c2("95  "), c1("jump: -1")},
			duration: 400 * time.Millisecond,
		},
		{
			caption:  "a negative number is relative the the last offset",
			template: tpl,
			args:     []interface{}{c2("40  "), c2("95  "), c1("jump: -10")},
			duration: 400 * time.Millisecond,
		},
		{
			caption:  "a negative number is relative the the last offset",
			template: tpl,
			args:     []interface{}{c2("40  "), c2("95  "), c1("jump: -100")},
			duration: 400 * time.Millisecond,
		},
		{
			caption:  "a negative number is relative the the last offset",
			template: tpl,
			args:     []interface{}{fmt.Sprintf(c2("%s  "), c3("40")), fmt.Sprintf(c2("%s  "), c3("95")), c1("jump: -100")},
			duration: 1000 * time.Millisecond,
		},
		//	rows := c2(`0             0                      %s                    1100                   1100
		//1             55                     %s                    2405                   2320`)
		{
			caption:  "a negative number is relative the the last offset",
			template: tpl,
			args:     []interface{}{fmt.Sprintf(c2("%s"), c3("1000")), fmt.Sprintf(c2("%s"), c3("2220")), c1("")},
			duration: 1000 * time.Millisecond,
		},
		{
			caption:  "a negative number is relative the the last offset",
			template: tpl,
			args:     []interface{}{c2("1000"), c2("2220"), c1("")},
			duration: 2000 * time.Millisecond,
		},
	}
}
