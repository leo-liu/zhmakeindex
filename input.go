package main

import (
	"bufio"
	"errors"
	"io"
	"log"
	"os"
	"unicode"
)

type InputIndex []IndexEntry

func NewInputIndex(option *Options, style *InputStyle) *InputIndex {
	in := InputIndex{}
	for _, idxname := range option.input {
		idxfile, err := os.Open(idxname)
		if err != nil {
			log.Fatalln(err.Error())
		}
		defer idxfile.Close()
		idxreader := bufio.NewReader(idxfile)
		for {
			entry, err := ScanIndexEntry(idxreader, style)
			if err == io.EOF {
				break
			} else if err == ScanSyntaxError {
				log.Println(err.Error())
				// 跳过一行
				for isprefix := true; isprefix; {
					var err error
					_, isprefix, err = idxreader.ReadLine()
					if err != nil {
						break
					}
				}
			} else if err != nil {
				log.Fatalln(err.Error())
			} else {
				in = append(in, *entry)
			}
		}
	}
	return &in
}

func skipspaces(reader *bufio.Reader) error {
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

func ScanIndexEntry(reader *bufio.Reader, style *InputStyle) (*IndexEntry, error) {
	entry := IndexEntry{}
	// 跳过空白符
	if err := skipspaces(reader); err != nil {
		return nil, err
	}
	// 跳过 keyword
	for i := 0; i < len(style.keyword); i++ {
		c, err := reader.ReadByte()
		if err != nil {
			return nil, err
		}
		if c != style.keyword[i] {
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
	token := []rune{}
L_scan_kv:
	for {
		r, _, err := reader.ReadRune()
		if err != nil {
			return nil, err
		}
		// debug.Printf("字符 %c, 状态 %d\n", r, state) //// DEBUG only
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
				token = []rune{}
				state = next
			}
			if quoted {
				token = append(token, r)
				quoted = false
			} else if r == style.arg_open {
				return nil, ScanSyntaxError
			} else if r == style.arg_close {
				pushtoken(0)
				break L_scan_kv
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
				token = []rune{}
				state = next
			}
			if quoted {
				token = append(token, r)
				quoted = false
			} else if r == style.arg_open || r == style.actual {
				return nil, ScanSyntaxError
			} else if r == style.arg_close {
				pushtoken(0)
				break L_scan_kv
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
			} else if r == style.arg_open || r == style.actual || r == style.encap || r == style.level {
				return nil, ScanSyntaxError
			} else if r == style.arg_close {
				break L_scan_kv
			} else if r == style.range_open {
				entry.pagerange = PAGE_OPEN
			} else if r == style.range_close {
				entry.pagerange = PAGE_CLOSE
			} else if r == style.quote {
				quoted = true
			} else {
				entry.pagerange = PAGE_NORMAL
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
			} else if r == style.arg_open || r == style.actual || r == style.encap || r == style.level {
				return nil, ScanSyntaxError
			} else if r == style.arg_close {
				entry.pagefmt = string(token)
				break L_scan_kv
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
	token = []rune{}
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
				entry.page = string(token)
				break L_scan_page
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
	level     []IndexEntryKV
	pagefmt   string
	page      string
	pagerange PageRange
}

type PageRange int

const (
	PAGE_NORMAL PageRange = iota
	PAGE_OPEN
	PAGE_CLOSE
)

type IndexEntryKV struct {
	key  string
	text string
}

//func (ind InputIndex) Len() int {
//	return len(ind)
//}

//func (ind InputIndex) Swap(i, j int) {
//	ind[i], ind[j] = ind[j], ind[i]
//}

//func (ind InputIndex) Less(i, j int) bool {
//	return true
//}
