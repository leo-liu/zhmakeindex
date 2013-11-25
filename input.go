package main

import (
	"errors"
	"fmt"
	"github.com/yasushi-saito/rbtree"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"unicode"
)

type InputIndex []IndexEntry

func NewInputIndex(option *InputOptions, style *InputStyle) *InputIndex {
	inset := rbtree.NewTree(CompareIndexEntry)

	if option.stdin {
		readIdxFile(inset, os.Stdin, style)
	} else {
		for _, idxname := range option.input {
			idxfile, err := os.Open(idxname)
			if err != nil {
				log.Fatalln(err.Error())
			}
			readIdxFile(inset, idxfile, style)
			idxfile.Close()
		}
	}

	var in InputIndex
	for iter := inset.Min(); !iter.Limit(); iter = iter.Next() {
		pentry := iter.Item().(*IndexEntry)
		in = append(in, *pentry)
	}
	return &in
}

func readIdxFile(inset *rbtree.Tree, idxfile *os.File, style *InputStyle) {
	log.Printf("读取输入文件 %s ……\n", idxfile.Name())
	accepted, rejected := 0, 0
	idxreader := NewNumberdReader(idxfile)
	for {
		entry, err := ScanIndexEntry(idxreader, style)
		if err == io.EOF {
			break
		} else if err == ScanSyntaxError {
			rejected++
			log.Printf("%s:%d: %s\n", idxfile.Name(), idxreader.Line(), err.Error())
			// 跳过一行
			if err := idxreader.SkipLine(); err == io.EOF {
				break
			} else if err != nil {
				log.Fatalln(err.Error())
			}
		} else if err != nil {
			log.Fatalln(err.Error())
		} else {
			accepted++
			if old := inset.Get(entry); old != nil {
				oldentry := old.(*IndexEntry)
				oldentry.pagelist = append(oldentry.pagelist, entry.pagelist...)
			} else {
				// entry 不在集合 inset 中时，插入 entry 本身和所有祖先节点，祖先不含页码
				for len(entry.level) > 0 {
					inset.Insert(entry)
					parent := &IndexEntry{
						level:    entry.level[:len(entry.level)-1],
						pagelist: nil,
					}
					if inset.Get(parent) != nil {
						break
					} else {
						entry = parent
					}
				}
			}
		}
	}
	log.Printf("接受 %d 项，拒绝 %d 项。\n", accepted, rejected)
}

func skipspaces(reader *NumberdReader) error {
	for {
		r, _, err := reader.ReadRune()
		if err != nil {
			return err
		} else if !unicode.IsSpace(r) {
			reader.UnreadRune()
			break
		}
	}
	return nil
}

func ScanIndexEntry(reader *NumberdReader, style *InputStyle) (*IndexEntry, error) {
	var entry IndexEntry
	entry.pagelist = make([]PageInput, 1)
	// 跳过空白符
	if err := skipspaces(reader); err != nil {
		return nil, err
	}
	// 跳过 keyword
	for _, r := range style.keyword {
		new_r, _, err := reader.ReadRune()
		if err != nil {
			return nil, err
		}
		if new_r != r {
			return nil, ScanSyntaxError
		}
	}
	// 自动机状态
	const (
		SCAN_OPEN = iota
		SCAN_KEY
		SCAN_VALUE
		SCAN_COMMAND
		SCAN_PAGE
		SCAN_PAGERANGE
	)
	// 从 arg_open 开始扫描到 arg_close，处理索引项
	state := SCAN_OPEN
	quoted := false
	escaped := false
	arg_depth := 0
	var token []rune
	entry.pagelist[0].rangetype = PAGE_NORMAL
L_scan_kv:
	for {
		r, _, err := reader.ReadRune()
		if err != nil {
			return nil, err
		}
		//debug.Printf("字符 %2c, 状态 %d, quoted %5v, escaped %5v, arg_depth %d\n", r, state, quoted, escaped, arg_depth) //// DEBUG only
		switch state {
		case SCAN_OPEN:
			if !quoted && r == style.arg_open {
				state = SCAN_KEY
			} else {
				return nil, ScanSyntaxError
			}
		case SCAN_KEY:
			pushtoken := func(next int) {
				str := string(token)
				entry.level = append(entry.level, IndexEntryKV{key: str, text: str})
				token = nil
				state = next
			}
			if quoted {
				token = append(token, r)
				quoted = false
				break
			} else if r == style.arg_open {
				token = append(token, r)
				arg_depth++
			} else if r == style.arg_close {
				if arg_depth == 0 {
					pushtoken(0)
					break L_scan_kv
				} else {
					token = append(token, r)
					arg_depth--
				}
			} else if r == style.actual {
				pushtoken(SCAN_VALUE)
			} else if r == style.encap {
				pushtoken(SCAN_PAGERANGE)
			} else if r == style.level {
				pushtoken(SCAN_KEY)
			} else if r == style.quote && !escaped {
				quoted = true
			} else {
				token = append(token, r)
			}
			if r == style.escape {
				escaped = true
			} else {
				escaped = false
			}
		case SCAN_VALUE:
			pushtoken := func(next int) {
				str := string(token)
				entry.level[len(entry.level)-1].text = str
				token = nil
				state = next
			}
			if quoted {
				token = append(token, r)
				quoted = false
				break
			} else if r == style.actual {
				return nil, ScanSyntaxError
			} else if r == style.arg_open {
				token = append(token, r)
				arg_depth++
			} else if r == style.arg_close {
				if arg_depth == 0 {
					pushtoken(0)
					break L_scan_kv
				} else {
					token = append(token, r)
					arg_depth--
				}
			} else if r == style.encap {
				pushtoken(SCAN_PAGERANGE)
			} else if r == style.level {
				pushtoken(SCAN_KEY)
			} else if r == style.quote && !escaped {
				quoted = true
			} else {
				token = append(token, r)
			}
			if r == style.escape {
				escaped = true
			} else {
				escaped = false
			}
		case SCAN_PAGERANGE:
			if quoted {
				token = append(token, r)
				quoted = false
				break
			} else if r == style.arg_open || r == style.arg_close || r == style.actual || r == style.encap || r == style.level {
				// 注意 encap 符号后不能直接加 arg_open、arg_close 等符号
				return nil, ScanSyntaxError
			} else if r == style.range_open {
				entry.pagelist[0].rangetype = PAGE_OPEN
			} else if r == style.range_close {
				entry.pagelist[0].rangetype = PAGE_CLOSE
			} else if r == style.quote {
				quoted = true
			} else {
				token = append(token, r)
			}
			state = SCAN_COMMAND
			if r == style.escape {
				escaped = true
			} else {
				escaped = false
			}
		case SCAN_COMMAND:
			if quoted {
				token = append(token, r)
				quoted = false
				break
			} else if r == style.actual || r == style.encap || r == style.level {
				return nil, ScanSyntaxError
			} else if r == style.arg_open {
				token = append(token, r)
				arg_depth++
			} else if r == style.arg_close {
				if arg_depth == 0 {
					entry.pagelist[0].encap = string(token)
					break L_scan_kv
				} else {
					token = append(token, r)
					arg_depth--
				}
			} else if r == style.quote && !escaped {
				quoted = true
			} else {
				token = append(token, r)
			}
			if r == style.escape {
				escaped = true
			} else {
				escaped = false
			}
		default:
			log.Fatalln("内部错误")
		}
	}
	// 跳过空白符
	if err := skipspaces(reader); err != nil {
		return nil, err
	}
	// 从 arg_open 开始扫描到 arg_close，处理页码
	state = SCAN_OPEN
	token = nil
L_scan_page:
	for {
		r, _, err := reader.ReadRune()
		if err != nil {
			return nil, err
		}
		// debug.Printf("字符 %c, 状态 %d\n", r, state) //// DEBUG only
		switch state {
		case SCAN_OPEN:
			if r == style.arg_open {
				state = SCAN_PAGE
			} else {
				return nil, ScanSyntaxError
			}
		case SCAN_PAGE:
			if r == style.arg_close {
				entry.pagelist[0].format, entry.pagelist[0].page, err = scanNumber(token)
				if err != nil {
					return nil, err
				}
				break L_scan_page
			} else if r == style.arg_open {
				return nil, ScanSyntaxError
			} else {
				token = append(token, r)
			}
		default:
			log.Fatalln("内部错误")
		}
		// 未实现对 style.page_compositor 的处理
	}
	// debug.Println(entry) //// DEBUG only
	return &entry, nil
}

var ScanSyntaxError = errors.New("索引项语法错误")

type IndexEntry struct {
	level    []IndexEntryKV
	pagelist []PageInput
}

// 实现 rbtree.CompareFunc
func CompareIndexEntry(a, b rbtree.Item) int {
	x := a.(*IndexEntry)
	y := b.(*IndexEntry)
	for i := range x.level {
		if i >= len(y.level) {
			return 1
		}
		if x.level[i].key < y.level[i].key {
			return -1
		} else if x.level[i].key > y.level[i].key {
			return 1
		}
		if x.level[i].text < y.level[i].text {
			return -1
		} else if x.level[i].text > y.level[i].text {
			return 1
		}
	}
	if len(x.level) < len(y.level) {
		return -1
	}
	return 0
}

type IndexEntryKV struct {
	key  string
	text string
}

type PageInput struct {
	page      int
	format    NumFormat
	encap     string
	rangetype RangeType
}

func (p *PageInput) NumString() string {
	return p.format.String(p.page)
}

type NumFormat int

const (
	NUM_ARABIC NumFormat = iota
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
func (numfmt NumFormat) String(num int) string {
	switch numfmt {
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
		return ""
	}
}

func romanNumString(num int, upper bool) string {
	if num < 1 {
		return ""
	}
	var numstr []rune
	for num > 1000 {
		numstr = append(numstr, 'm')
		num -= 1000
	}
	if num >= 900 {
		numstr = append(numstr, 'c', 'm')
		num -= 900
	}
	if num >= 500 {
		numstr = append(numstr, 'd')
		num -= 500
	}
	if num >= 400 {
		numstr = append(numstr, 'c', 'd')
		num -= 400
	}
	for num >= 100 {
		numstr = append(numstr, 'c')
		num -= 100
	}
	if num >= 90 {
		numstr = append(numstr, 'x', 'c')
		num -= 90
	}
	if num >= 50 {
		numstr = append(numstr, 'l')
		num -= 50
	}
	if num >= 40 {
		numstr = append(numstr, 'x', 'l')
		num -= 40
	}
	for num >= 10 {
		numstr = append(numstr, 'x')
		num -= 10
	}
	if num >= 9 {
		numstr = append(numstr, 'i', 'x')
		num -= 9
	}
	if num >= 5 {
		numstr = append(numstr, 'v')
		num -= 5
	}
	if num >= 4 {
		numstr = append(numstr, 'i', 'v')
		num -= 4
	}
	for num >= 1 {
		numstr = append(numstr, 'i')
		num--
	}
	if upper {
		return strings.ToUpper(string(numstr))
	} else {
		return string(numstr)
	}
}

type RangeType int

const (
	PAGE_OPEN RangeType = iota
	PAGE_NORMAL
	PAGE_CLOSE
)
