package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/yasushi-saito/rbtree"
	"io"
	"log"
	"os"
	"strconv"
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

type NumberdReader struct {
	reader   *bufio.Reader
	line     int
	lastRune rune
}

func NewNumberdReader(reader io.Reader) *NumberdReader {
	return &NumberdReader{
		reader:   bufio.NewReader(reader),
		line:     1,
		lastRune: -1,
	}
}

func (reader *NumberdReader) ReadRune() (r rune, size int, err error) {
	r, size, err = reader.reader.ReadRune()
	reader.lastRune = r
	if r == '\n' {
		reader.line++
	}
	return r, size, err
}

func (reader *NumberdReader) UnreadRune() error {
	err := reader.reader.UnreadRune()
	if err == nil {
		if reader.lastRune == '\n' {
			reader.line--
		}
		reader.lastRune = -1
	}
	return err
}

func (reader *NumberdReader) SkipLine() error {
	reader.line++
	_, err := reader.reader.ReadBytes('\n')
	return err
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
			log.Printf("%s:%d: %s\n", idxfile.Name(), idxreader.line, err.Error())
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
				entry.pagelist[0].format, entry.pagelist[0].page = scanNumber(token)
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

// 未完整实现，目前仅有阿拉伯数字
func scanNumber(token []rune) (NumFormat, int) {
	return NUM_ARABIC, scanArabic(token)
}

func scanArabic(token []rune) int {
	num, _ := strconv.Atoi(string(token))
	return num
}

// 未完整实现，仅阿拉伯数字
func (fmt NumFormat) String(num int) string {
	return arabicNumString(num)
}

func arabicNumString(num int) string {
	return fmt.Sprint(num)
}

type RangeType int

const (
	PAGE_OPEN RangeType = iota
	PAGE_NORMAL
	PAGE_CLOSE
)
