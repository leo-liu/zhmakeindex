// $Id: pagenumber.go,v 76b101661244 2014/08/20 16:59:14 leoliu $

package main

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

type PageNumber struct {
	page      int
	format    NumFormat
	encap     string
	rangetype RangeType
}

func (p *PageNumber) Empty() PageNumber {
	return PageNumber{
		page: 0, format: NUM_UNKNOWN, encap: p.encap, rangetype: PAGE_UNKNOWN,
	}
}

func (p PageNumber) String() string {
	return p.format.Format(p.page)
}

type NumFormat int

const (
	NUM_UNKNOWN NumFormat = iota
	NUM_ARABIC
	NUM_ROMAN_LOWER
	NUM_ROMAN_UPPER
	NUM_ALPH_LOWER
	NUM_ALPH_UPPER
)

// 读入数字，并粗略判断其格式
func scanNumber(token []rune) (NumFormat, int, error) {
	if len(token) == 0 {
		return 0, 0, ScanSyntaxError
	}
	if r := token[0]; unicode.IsDigit(r) {
		num, err := scanArabic(token)
		return NUM_ARABIC, num, err
	} else if RomanLowerValue[r] != 0 {
		num, err := scanRomanLower(token)
		return NUM_ROMAN_LOWER, num, err
	} else if RomanUpperValue[r] != 0 {
		num, err := scanRomanUpper(token)
		return NUM_ROMAN_UPPER, num, err
	} else if 'a' <= r && r <= 'z' {
		num, err := scanAlphLower(token)
		return NUM_ALPH_LOWER, num, err
	} else if 'A' <= r && r <= 'Z' {
		num, err := scanAlphUpper(token)
		return NUM_ALPH_UPPER, num, err
	}
	return 0, 0, ScanSyntaxError
}

func scanArabic(token []rune) (int, error) {
	num, err := strconv.Atoi(string(token))
	if err != nil {
		err = ScanSyntaxError
	}
	return num, err
}

func scanRomanLower(token []rune) (int, error) {
	return scanRoman(token, RomanLowerValue)
}

func scanRomanUpper(token []rune) (int, error) {
	return scanRoman(token, RomanUpperValue)
}

func scanRoman(token []rune, romantable map[rune]int) (int, error) {
	num := 0
	for i, r := range token {
		if romantable[r] == 0 {
			return 0, ScanSyntaxError
		}
		if i == 0 || romantable[r] <= romantable[token[i-1]] {
			num += romantable[r]
		} else {
			num += romantable[r] - 2*romantable[token[i-1]]
		}
	}
	return num, nil
}

var RomanLowerValue = map[rune]int{
	'i': 1, 'v': 5, 'x': 10, 'l': 50, 'c': 100, 'd': 500, 'm': 1000,
}
var RomanUpperValue = map[rune]int{
	'I': 1, 'V': 5, 'X': 10, 'L': 50, 'C': 100, 'D': 500, 'M': 1000,
}

func scanAlphLower(token []rune) (int, error) {
	if len(token) != 1 || token[0] < 'a' || token[0] > 'z' {
		return 0, ScanSyntaxError
	}
	return int(token[0]-'a') + 1, nil
}

func scanAlphUpper(token []rune) (int, error) {
	if len(token) != 1 || token[0] < 'A' || token[0] > 'Z' {
		return 0, ScanSyntaxError
	}
	return int(token[0]-'A') + 1, nil
}

// 按格式输出数字
func (numfmt NumFormat) Format(num int) string {
	switch numfmt {
	case NUM_UNKNOWN:
		return "?"
	case NUM_ARABIC:
		return fmt.Sprint(num)
	case NUM_ALPH_LOWER:
		return string('a' + num)
	case NUM_ALPH_UPPER:
		return string('A' + num)
	case NUM_ROMAN_LOWER:
		return romanNumString(num, false)
	case NUM_ROMAN_UPPER:
		return romanNumString(num, true)
	default:
		panic("数字格式错误")
	}
}

func romanNumString(num int, upper bool) string {
	if num < 1 {
		return ""
	}
	type pair struct {
		symbol string
		value  int
	}
	var romanTable = []pair{
		{"m", 1000}, {"cm", 900}, {"d", 500}, {"cd", 400}, {"c", 100}, {"xc", 90},
		{"l", 50}, {"xl", 40}, {"x", 10}, {"ix", 9}, {"v", 5}, {"iv", 4}, {"i", 1},
	}
	var numstr []rune
	for _, p := range romanTable {
		for num > p.value {
			numstr = append(numstr, []rune(p.symbol)...)
			num -= p.value
		}
	}
	if upper {
		return strings.ToUpper(string(numstr))
	} else {
		return string(numstr)
	}
}
