// $Id$

package main

import (
	"fmt"
	"unicode"
	"unicode/utf8"

	"code.google.com/p/zhmakeindex/CJK"
)

// 汉字按部首-除部首笔画数排序，汉字按部首分组排在英文字母组后面
type RadicalIndexCollator struct{}

func (_ RadicalIndexCollator) InitGroups(style *OutputStyle) []IndexGroup {
	// 分组：符号、数字、字母 A..Z
	groups := make([]IndexGroup, 2+26+CJK.MAX_RADICAL)
	if style.headings_flag > 0 {
		groups[0].name = style.symhead_positive
		groups[1].name = style.numhead_positive
		for alph, i := 'A', 2; alph <= 'Z'; alph++ {
			groups[i].name = string(alph)
			i++
		}
	} else if style.headings_flag < 0 {
		groups[0].name = style.symhead_negative
		groups[1].name = style.numhead_negative
		for alph, i := 'a', 2; alph <= 'z'; alph++ {
			groups[i].name = string(alph)
			i++
		}
	}
	for r, i := 1, 2+26; r < CJK.MAX_RADICAL+1; r++ {
		var radicalName string
		if CJK.Radicals[r].Simplified != 0 && style.radical_simplified_flag != 0 {
			radicalName = fmt.Sprintf("%c%s%c%s",
				CJK.Radicals[r].Origin, style.radical_simplified_prefix, CJK.Radicals[r].Simplified, style.radical_simplified_suffix)
		} else {
			radicalName = string(CJK.Radicals[r].Origin)
		}
		groups[i].name = style.radical_prefix + radicalName + style.radical_suffix
		i++
	}
	return groups
}

// 取得分组
func (_ RadicalIndexCollator) Group(entry *IndexEntry) int {
	first, _ := utf8.DecodeRuneInString(entry.level[0].key)
	first = unicode.ToLower(first)
	switch {
	case IsNumString(entry.level[0].key):
		return 1
	case 'a' <= first && first <= 'z':
		return 2 + int(first) - 'a'
	case CJK.RadicalStrokes[first] != "":
		// 首字部首
		return 2 + 26 + (CJK.RadicalStrokes[first].Radical() - 1)
	default:
		// 符号组
		return 0
	}
}

// 按汉字部首、除部首笔画数序比较两个字符大小
func (_ RadicalIndexCollator) RuneCmp(a, b rune) int {
	a_rs, b_rs := CJK.RadicalStrokes[a], CJK.RadicalStrokes[b]
	switch {
	case a_rs == "" && b_rs == "":
		return RuneCmpIgnoreCases(a, b)
	case a_rs == "" && b_rs != "":
		return -1
	case a_rs != "" && b_rs == "":
		return 1
	case a_rs < b_rs:
		return -1
	case a_rs > b_rs:
		return 1
	default:
		return int(a - b)
	}
}

// 判断是否字母或汉字
func (_ RadicalIndexCollator) IsLetter(r rune) bool {
	r = unicode.ToLower(r)
	switch {
	case 'a' <= r && r <= 'z':
		return true
	case CJK.RadicalStrokes[r] != "":
		return true
	default:
		return false
	}
}
