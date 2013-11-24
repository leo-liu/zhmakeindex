// reading_collator.go
package main

import (
	"unicode"
)

// 汉字按拼音排序，按拼音首字母与英文一起分组
type ReadingIndexCollator struct{}

func (_ ReadingIndexCollator) InitGroups(style *OutputStyle) []IndexGroup {
	// 分组：数字、符号、字母 A..Z
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
	return groups
}

// 取得分组
func (_ ReadingIndexCollator) Group(entry *IndexEntry) int {
	first := ([]rune(entry.level[0].key))[0]
	first = unicode.ToLower(first)
	switch {
	case unicode.IsNumber(first):
		return 0
	case 'a' <= first && first <= 'z':
		return 2 + int(first) - 'a'
	case CJKreadings[first] != "":
		// 拼音首字母
		reading_first := int(CJKreadings[first][0])
		return 2 + reading_first - 'a'
	default:
		// 符号组
		return 1
	}
	return 0
}

// 按汉字读音比较两个串的大小
func (_ ReadingIndexCollator) Strcmp(a, b string) int {
	a_rune, b_rune := []rune(a), []rune(b)
	for i := range a_rune {
		if i >= len(b_rune) {
			return 1
		}
		cmp := runecmpByReading(a_rune[i], b_rune[i])
		if cmp != 0 { // 读音不同
			return cmp
		} else if a_rune[i] != b_rune[i] { // 读音相同、字符不同
			return int(a_rune[i] - b_rune[i])
		}
	}
	if len(a_rune) < len(b_rune) {
		return -1
	}
	return 0
}

// 按汉字读音比较两个字符
func runecmpByReading(a, b rune) int {
	a_reading, b_reading := CJKreadings[a], CJKreadings[b]
	switch {
	case a_reading == "" && b_reading == "":
		return int(a - b)
	case a_reading == "" && b_reading != "":
		return -1
	case a_reading != "" && b_reading == "":
		return 1
	default:
		if a_reading < b_reading {
			return -1
		} else if a_reading > b_reading {
			return 1
		} else {
			return 0
		}
	}
}
