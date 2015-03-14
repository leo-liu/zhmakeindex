package main

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// 最大的整数
const MaxInt = int(^uint(0) >> 1)

type Page struct {
	numbers    []PageNumber
	compositor string
	encap      string
	rangetype  RangeType
}

// 按 p 生成一个与之输出类型相同的空页码
func (p *Page) Empty() *Page {
	return &Page{
		numbers:    nil,
		compositor: p.compositor,
		encap:      p.encap,
		rangetype:  PAGE_UNKNOWN,
	}
}

func (p *Page) String() string {
	var page_str []string
	for _, pn := range p.numbers {
		page_str = append(page_str, pn.String())
	}
	return strings.Join(page_str, p.compositor)
}

// 判断两个页码是否类型一致
// 两个页码类型一致指它们有相同多个类型相同的数字构成
func (page *Page) Compatible(other *Page) bool {
	if len(page.numbers) != len(other.numbers) {
		return false
	}
	for i := 0; i < len(page.numbers); i++ {
		if page.numbers[i].format != other.numbers[i].format {
			return false
		}
	}
	return true
}

// 求两个页码 page 与 other 之差（绝对值）
// 如果不一致，返回 -1；如果不是最后一段数字不同，返回 MaxInt
func (page *Page) Diff(other *Page) int {
	if !page.Compatible(other) {
		return -1
	}
	depth := len(page.numbers)
	for i := 0; i < depth-1; i++ {
		if page.numbers[i].num != other.numbers[i].num {
			return MaxInt
		}
	}
	abs := func(x int) int {
		if x >= 0 {
			return x
		} else {
			return -x
		}
	}
	return abs(page.numbers[depth-1].num - other.numbers[depth-1].num)
}

// 按字典序比较两个页码数字串的大小，返回负、零、正值
// 不同类型的序关系由参数 precedence 给出，不一致的页码仍有大小关系
// 不比较页码的 encap、rangetype 信息
func (page *Page) Cmp(other *Page, precedence map[NumFormat]int) int {
	for i := 0; i < len(page.numbers) && i < len(other.numbers); i++ {
		a, b := page.numbers[i], other.numbers[i]
		if precedence[a.format] != precedence[b.format] {
			return precedence[a.format] - precedence[b.format]
		} else if a.num != b.num {
			return a.num - b.num
		}
	}
	if len(page.numbers) != len(other.numbers) {
		return len(page.numbers) - len(other.numbers)
	}
	return 0
}

type PageNumber struct {
	format NumFormat
	num    int
}

func (p PageNumber) String() string {
	return p.format.Format(p.num)
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

// 将字符串解析为一串页码数字
func scanPage(token []rune, compositor string) ([]PageNumber, error) {
	numstr_list := strings.Split(string(token), compositor)
	var nums []PageNumber
	for _, numstr := range numstr_list {
		pn, err := scanNumber([]rune(numstr))
		if err != nil {
			return nil, err
		}
		nums = append(nums, pn)
	}
	return nums, nil
}

// 读入数字，并粗略判断其格式
func scanNumber(token []rune) (PageNumber, error) {
	if len(token) == 0 {
		return PageNumber{}, ScanSyntaxError
	}
	if r := token[0]; unicode.IsDigit(r) {
		num, err := scanArabic(token)
		return PageNumber{format: NUM_ARABIC, num: num}, err
	} else if romanLowerValue[r] != 0 {
		num, err := scanRomanLower(token)
		return PageNumber{format: NUM_ROMAN_LOWER, num: num}, err
	} else if romanUpperValue[r] != 0 {
		num, err := scanRomanUpper(token)
		return PageNumber{format: NUM_ROMAN_UPPER, num: num}, err
	} else if 'a' <= r && r <= 'z' {
		num, err := scanAlphLower(token)
		return PageNumber{format: NUM_ALPH_LOWER, num: num}, err
	} else if 'A' <= r && r <= 'Z' {
		num, err := scanAlphUpper(token)
		return PageNumber{format: NUM_ALPH_UPPER, num: num}, err
	}
	return PageNumber{}, ScanSyntaxError
}

func scanArabic(token []rune) (int, error) {
	num, err := strconv.Atoi(string(token))
	if err != nil {
		err = ScanSyntaxError
	}
	return num, err
}

func scanRomanLower(token []rune) (int, error) {
	return scanRoman(token, romanLowerValue)
}

func scanRomanUpper(token []rune) (int, error) {
	return scanRoman(token, romanUpperValue)
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

var romanLowerValue = map[rune]int{
	'i': 1, 'v': 5, 'x': 10, 'l': 50, 'c': 100, 'd': 500, 'm': 1000,
}
var romanUpperValue = map[rune]int{
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
		for num >= p.value {
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
