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
	partitionHeader := c1("offset       message    topic: items partition: 0 start: 0 end: 1017")
	partitionRows := c2(`0            Robert,Scott,46
1            Richard,Scott,58
2            Martha,Payne,13
3            Tabitha,Bass,71
4            Katherine,Robertson,83
5            Jane,Hawkins,82
6            Bob,Nash,5
7            Jacqueline,Scott,61
8            Dave,Pratt,37
9            Debra,Gross,63
10           Ashley,Gross,19`)

	partitionTpl := fmt.Sprintf(`%s
%s
%%s`, partitionHeader, partitionRows)

	partitionRows2 := c2(`25           Yolanda,Crawford,66
26           Donnie,Lopez,77
27           Grace,Bass,67
28           Abraham,Pratt,9
29           Bobbie,Cooper,56
30           Sabrina,Nash,47
31           Elias,Singleton,47
32           Sylvester,Cooper,16
33           Katherine,Scott,42
34           Bobbie,Steele,43
35           Donnie,Pratt,67`)

	partitionTpl2 := fmt.Sprintf(`%s
%s
%%s`, partitionHeader, partitionRows2)

	header := c1("partition     1st offset             current offset         last offset            size")
	rows := c2(`0             0                      %s                   1100                   1100
1             55                     %s                   2405                   2320`)
	tpl := fmt.Sprintf(`%s
%s









%%s`, header, rows)

	return []helpStep{
		{
			caption:  "when jumping on a partition, the number you",
			template: partitionTpl,
			args:     []interface{}{""},
			duration: 500 * time.Millisecond,
		},
		{
			caption:  "when jumping on a partition, the number you",
			template: partitionTpl,
			args:     []interface{}{c1("jump: ")},
			duration: 500 * time.Millisecond,
		},
		{
			caption:  "when jumping on a partition, the number you",
			template: partitionTpl,
			args:     []interface{}{c1("jump: 2")},
			duration: 500 * time.Millisecond,
		},
		{
			caption:  "when jumping on a partition, the number you",
			template: partitionTpl,
			args:     []interface{}{c1("jump: 25")},
			duration: 500 * time.Millisecond,
		},
		{
			caption:  "enter sets the current offset",
			template: partitionTpl,
			args:     []interface{}{c1("jump: 25")},
			duration: 500 * time.Millisecond,
		},
		{
			caption:  "enter sets the current offset",
			template: partitionTpl2,
			args:     []interface{}{""},
			duration: 4000 * time.Millisecond,
		},
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
