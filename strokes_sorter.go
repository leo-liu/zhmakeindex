// strokes_sorter.go
package main

import (
	"code.google.com/p/go.text/collate"
	"code.google.com/p/go.text/language"
	"sort"
	"strconv"
	"unicode"
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

	// 先整体排序
	sort.Sort(IndexEntrySliceByStroke{*input})

	// 再依次分组添加
	pagesorter := NewPageSorter(style)
	for _, entry := range *input {
		pageranges := pagesorter.Sort(entry.pagelist)
		item := IndexItem{
			level: len(entry.level) - 1,
			text:  entry.level[len(entry.level)-1].text,
			page:  pageranges,
		}
		group := sorter.Group(&entry)
		out.groups[group].items = append(out.groups[group].items, item)
	}

	return out
}

type IndexEntrySliceByStroke struct {
	entries []IndexEntry
}

func (s IndexEntrySliceByStroke) Len() int {
	return len(s.entries)
}

func (s IndexEntrySliceByStroke) Swap(i, j int) {
	s.entries[i], s.entries[j] = s.entries[j], s.entries[i]
}

func (s IndexEntrySliceByStroke) Less(i, j int) bool {
	a, b := s.entries[i], s.entries[j]
	for i := range a.level {
		if i >= len(b.level) {
			return false
		}
		//debug.Println(i, a.level[i], b.level[i])
		keycmp := CmpstrByStroke(a.level[i].key, b.level[i].key)
		if keycmp < 0 {
			return true
		} else if keycmp > 0 {
			return false
		}
		textcmp := CmpstrByStroke(a.level[i].text, b.level[i].text)
		if textcmp < 0 {
			return true
		} else if textcmp > 0 {
			return false
		}
	}
	if len(a.level) < len(b.level) {
		return true
	}
	return false
}

var CollatorByStroke = collate.New(language.Make("zh_stroke"))

// 按笔划序比较两个串的大小
func CmpstrByStroke(s, t string) int {
	return CollatorByStroke.CompareString(s, t)
}

// 取得分组
func (sorter *StrokesSorter) Group(entry *IndexEntry) int {
	first := ([]rune(entry.level[0].key))[0]
	first = unicode.ToLower(first)
	switch {
	case unicode.IsNumber(first):
		return 0
	case 'a' <= first && first <= 'z':
		return 2 + int(first) - 'a'
	case CJKstrokes[first] > 0:
		return 2 + 26 + (CJKstrokes[first] - 1)
	default:
		// 符号组
		return 1
	}
}
