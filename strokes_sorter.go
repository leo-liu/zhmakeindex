// strokes_sorter.go
package main

import (
	"code.google.com/p/go.text/collate"
	"code.google.com/p/go.text/language"
	"strconv"
	"unicode"
)

type StrokesIndexCollator struct{}

func (_ StrokesIndexCollator) InitGroups(style *OutputStyle) []IndexGroup {
	// 分组：数字、符号、字母 A..Z、笔划 1..MAX_STROKE
	groups := make([]IndexGroup, 2+26+MAX_STROKE)
	if style.headings_flag > 0 {
		groups[0].name = style.numhead_positive
		groups[1].name = style.symhead_positive
		for alph, i := 'A', 2; alph <= 'Z'; alph++ {
			groups[i].name = string(alph)
			i++
		}
	} else if style.headings_flag < 0 {
		groups[0].name = style.numhead_negative
		groups[1].name = style.symhead_negative
		for alph, i := 'a', 2; alph <= 'z'; alph++ {
			groups[i].name = string(alph)
			i++
		}
	}
	for stroke, i := 1, 2+26; stroke <= MAX_STROKE; stroke++ {
		groups[i].name = strconv.Itoa(stroke) + " 划"
		i++
	}
	return groups
}

// 取得分组
func (_ StrokesIndexCollator) Group(entry *IndexEntry) int {
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

var CollatorByStroke = collate.New(language.Make("zh_stroke"))

// 按笔划序比较两个串的大小
func (_ StrokesIndexCollator) Strcmp(a, b string) int {
	return CollatorByStroke.CompareString(a, b)
}
