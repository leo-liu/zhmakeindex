// $Id: stroke_collator.go,v 440c8c98f478 2014/02/11 19:12:02 LeoLiu $

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
		groups[i].name = style.stroke_prefix + strconv.Itoa(stroke) + style.stroke_suffix
		i++
	}
	return groups
}

// 取得分组
func (_ StrokeIndexCollator) Group(entry *IndexEntry) int {
	first := ([]rune(entry.level[0].key))[0]
	first = unicode.ToLower(first)
	switch {
	case IsNumString(entry.level[0].key):
		return 0
	case 'a' <= first && first <= 'z':
		return 2 + int(first) - 'a'
	case len(CJKstrokes[first]) > 0:
		return 2 + 26 + (len(CJKstrokes[first]) - 1)
	default:
		// 符号组
		return 1
	}
}

// 按汉字笔画、笔顺序比较两个字符大小
// 笔画数不同的，短的在前；笔画数相同的，笔顺字典序；笔顺相同的，内码序
func (_ StrokeIndexCollator) RuneCmp(a, b rune) int {
	a_strokes, b_strokes := len(CJKstrokes[a]), len(CJKstrokes[b])
	switch {
	case a_strokes == 0 && b_strokes == 0:
		return RuneCmpIgnoreCases(a, b)
	case a_strokes == 0 && b_strokes != 0:
		return -1
	case a_strokes != 0 && b_strokes == 0:
		return 1
	case a_strokes != b_strokes:
		return a_strokes - b_strokes
	case CJKstrokes[a] < CJKstrokes[b]:
		return -1
	case CJKstrokes[a] > CJKstrokes[b]:
		return 1
	default:
		return int(a - b)
	}
}
