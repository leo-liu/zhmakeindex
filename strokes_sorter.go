// strokes_sorter.go
package main

import (
	"strconv"
)

type StrokesSorter struct{}

func (sorter *StrokesSorter) SortIndex(input *InputIndex, style *OutputStyle) *OutputIndex {
	out := new(OutputIndex)
	// 分组：数字、符号、字母 A..Z、笔划 1..MAX_STROKE
	out.groups = make([]IndexGroup, 2+26+MAX_STROKE)
	if style.headings_flag > 0 {
		out.groups[0].name = style.numhead_positive
		out.groups[1].name = style.symhead_positive
		for alph, i := 'A', 2; alph <= 'Z'; alph++ {
			out.groups[i].name = string(alph)
			i++
		}
	} else if style.headings_flag < 0 {
		out.groups[0].name = style.numhead_negative
		out.groups[1].name = style.symhead_negative
		for alph, i := 'a', 2; alph <= 'z'; alph++ {
			out.groups[i].name = string(alph)
			i++
		}
	}
	for stroke, i := 1, 2+26; stroke <= MAX_STROKE; stroke++ {
		out.groups[i].name = strconv.Itoa(stroke) + " 划"
		i++
	}
	out.groups[0].items = make([]IndexItem, 1)
	out.groups[0].items[0].text = "乙"
	out.groups[0].items[0].level = 0
	out.groups[0].items[0].page = []PageRange{{encap: "textit", begin: "1", end: "1"}}
	return out
}
