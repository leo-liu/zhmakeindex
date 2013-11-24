// stroke_collator.go
package main

import (
	"strconv"
	"unicode"
)

// 汉字按笔画排序，汉字按笔画分组排在英文字母组后面
type StrokeIndexCollator struct{}

func (_ StrokeIndexCollator) InitGroups(style *OutputStyle) []IndexGroup {
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
func (_ StrokeIndexCollator) Group(entry *IndexEntry) int {
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

// 按笔划序比较两个串的大小
func (_ StrokeIndexCollator) Strcmp(a, b string) int {
	a_rune, b_rune := []rune(a), []rune(b)
	for i := range a_rune {
		if i >= len(b_rune) {
			return 1
		}
		cmp := runecmpByStroke(a_rune[i], b_rune[i])
		if cmp != 0 { // 笔画数不同
			return cmp
		} else if a_rune[i] != b_rune[i] { // 笔画数相同、字符不同
			return int(a_rune[i] - b_rune[i])
		}
	}
	if len(a_rune) < len(b_rune) {
		return -1
	}
	return 0
}

// 按汉字笔划序比较两个字符大小
func runecmpByStroke(a, b rune) int {
	a_strokes, b_strokes := CJKstrokes[a], CJKstrokes[b]
	switch {
	case a_strokes == 0 && b_strokes == 0:
		return int(a - b)
	case a_strokes == 0 && b_strokes != 0:
		return -1
	case a_strokes != 0 && b_strokes == 0:
		return 1
	default:
		return a_strokes - b_strokes
	}
}
