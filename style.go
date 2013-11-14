package main

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// 输入格式
// 这里使用简单 struct 实现，需要大量分情况讨论。
// 也可以用 map 实现，代码可能会简短并易于扩展，但要动态处理类型
type InputStyle struct {
	keyword         string
	arg_open        rune
	arg_close       rune
	actual          rune
	encap           rune
	escape          rune
	level           rune
	quote           rune
	page_compositor string
	range_open      rune
	range_close     rune
}

func NewInputStyle() *InputStyle {
	in := &InputStyle{
		keyword:         "\\indexentry",
		arg_open:        '{',
		arg_close:       '}',
		actual:          '@',
		encap:           '|',
		escape:          '\\',
		level:           '!',
		quote:           '"',
		page_compositor: "-",
		range_open:      '(',
		range_close:     ')',
	}
	return in
}

type OutputStyle struct {
	preamble         string
	postamble        string
	setpage_prefix   string
	setpage_suffix   string
	group_skip       string
	headings_flag    int
	heading_prefix   string
	symhead_positive string
	symhead_negative string
	numhead_positive string
	numhead_negative string
	item_0           string
	item_1           string
	item_2           string
	item_01          string
	item_x1          string
	item_12          string
	item_x2          string
	delim_0          string
	delim_1          string
	delim_2          string
	delim_n          string
	delim_r          string
	delim_t          string
	encap_prefix     string
	encap_infix      string
	encap_suffix     string
	line_max         int
	indent_space     string
	indent_length    int
	suffix_2p        string
	suffix_3p        string
	suffix_mp        string
}

func NewOutputStyle() *OutputStyle {
	out := &OutputStyle{
		preamble:         "\\begin{theindex}\n",
		postamble:        "\n\n\\end{theindex}\n",
		setpage_prefix:   "\n  \\setcounter{page}{",
		setpage_suffix:   "}\n",
		group_skip:       "\n\n  \\indexspace\n",
		headings_flag:    0,
		heading_prefix:   "",
		symhead_positive: "Symbols",
		symhead_negative: "symbols",
		numhead_positive: "Numbers",
		numhead_negative: "numbers",
		item_0:           "\n  \\tiem ",
		item_1:           "\n    \\subitem ",
		item_2:           "\n      \\subsubitem ",
		item_01:          "\n    \\subitem ",
		item_x1:          "\n    \\subitem ",
		item_12:          "\n      \\subsubitem ",
		item_x2:          "\n      \\subsubitem ",
		delim_0:          ", ",
		delim_1:          ", ",
		delim_2:          ", ",
		delim_n:          ", ",
		delim_r:          "--",
		delim_t:          "",
		encap_prefix:     "\\",
		encap_infix:      "{",
		encap_suffix:     "}",
		line_max:         72,
		indent_space:     "\t\t",
		indent_length:    16,
		suffix_2p:        "",
		suffix_3p:        "",
		suffix_mp:        "",
	}
	return out
}

func NewStyles(option *Options) (*InputStyle, *OutputStyle) {
	in := NewInputStyle()
	out := NewOutputStyle()

	if option.style == "" {
		return in, out
	}
	// 读取格式文件，处理格式
	styleFile, err := os.Open(option.style)
	if err != nil {
		log.Fatalln(err.Error())
	}
	scanner := bufio.NewScanner(styleFile)
	scanner.Split(ScanStyleTokens)
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			log.Println(err.Error())
		}
		key := scanner.Text()
		if !scanner.Scan() {
			log.Println("格式文件不完整")
		}
		if err := scanner.Err(); err != nil {
			log.Println(err.Error())
		}
		value := scanner.Text()
		switch key {
		// 输入参数
		case "keyword":
			in.keyword = unquote(value)
		case "arg_open":
			in.arg_open = unquoteChar(value)
		case "arg_close":
			in.arg_close = unquoteChar(value)
		case "actual":
			in.actual = unquoteChar(value)
		case "encap":
			in.encap = unquoteChar(value)
		case "escape":
			in.escape = unquoteChar(value)
		case "level":
			in.level = unquoteChar(value)
		case "quote":
			in.quote = unquoteChar(value)
		case "page_compositor":
			in.page_compositor = unquote(value)
		case "range_open":
			in.range_open = unquoteChar(value)
		case "range_close":
			in.range_close = unquoteChar(value)
		// 输出参数
		case "preamble":
			out.preamble = unquote(value)
		case "postamble":
			out.postamble = unquote(value)
		case "setpage_prefix":
			out.setpage_prefix = unquote(value)
		case "setpage_suffix":
			out.setpage_suffix = unquote(value)
		case "group_skip":
			out.group_skip = unquote(value)
		case "headings_flag":
			out.headings_flag = parseInt(value)
		case "heading_prefix":
			out.heading_prefix = unquote(value)
		case "symhead_positive":
			out.symhead_positive = unquote(value)
		case "symhead_negative":
			out.symhead_negative = unquote(value)
		case "numhead_positive":
			out.numhead_positive = unquote(value)
		case "numhead_negative":
			out.numhead_negative = unquote(value)
		case "item_0":
			out.item_0 = unquote(value)
		case "item_1":
			out.item_1 = unquote(value)
		case "item_2":
			out.item_2 = unquote(value)
		case "item_01":
			out.item_01 = unquote(value)
		case "item_x1":
			out.item_x1 = unquote(value)
		case "item_12":
			out.item_12 = unquote(value)
		case "item_x2":
			out.item_x2 = unquote(value)
		case "delim_0":
			out.delim_0 = unquote(value)
		case "delim_1":
			out.delim_1 = unquote(value)
		case "delim_2":
			out.delim_2 = unquote(value)
		case "delim_n":
			out.delim_n = unquote(value)
		case "delim_r":
			out.delim_r = unquote(value)
		case "delim_t":
			out.delim_t = unquote(value)
		case "encap_prefix":
			out.encap_prefix = unquote(value)
		case "encap_infix":
			out.encap_infix = unquote(value)
		case "encap_suffix":
			out.encap_suffix = unquote(value)
		case "line_max":
			out.line_max = parseInt(value)
		case "indent_space":
			out.indent_space = unquote(value)
		case "indent_length":
			out.indent_length = parseInt(value)
		case "suffix_2p":
			out.suffix_2p = unquote(value)
		case "suffix_3p":
			out.suffix_3p = unquote(value)
		case "suffix_mp":
			out.suffix_mp = unquote(value)
		// 其他
		default:
			log.Printf("忽略未知格式 %s\n", key)
		}
	}
	return in, out
}

func unquote(src string) string {
	// 处理双引号中有换行符的串
	if src[0] == '"' {
		src = strings.Replace(src, "\n", "\\n", -1)
	}
	dst, err := strconv.Unquote(src)
	if err != nil {
		log.Println(err.Error())
	}
	return dst
}

func unquoteChar(src string) rune {
	src = unquote(src)
	dst, _, tail, err := strconv.UnquoteChar(src, 0)
	if tail != "" {
		err = strconv.ErrSyntax
	}
	if err != nil {
		log.Println(err.Error())
	}
	return dst
}

func parseInt(src string) int {
	i, err := strconv.ParseInt(src, 0, 0)
	if err != nil {
		log.Println(err.Error())
	}
	return int(i)
}

// bufio.SplitFunc 的实例
// 查找标识符、数字、单引号内的 rune、双引号或反引号内的串；跳过以 % 开头的注释（未完成）
// 实现参考了 bufio.ScanWords
func ScanStyleTokens(data []byte, atEOF bool) (advance int, token []byte, err error) {
	// 跳过空白和注释
	start := 0
	in_comment := false
	for width := 0; start < len(data); start += width {
		var r rune
		r, width = utf8.DecodeRune(data[start:])
		if !in_comment && r == '%' {
			in_comment = true
		} else if in_comment && r == '\n' {
			in_comment = false
		} else if !in_comment && !unicode.IsSpace(r) {
			break
		}
		// 其他情况跳过：注释中未遇到换行符，或者注释外遇到空白符
	}
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	// 首先读出第一个字符，按不同类型扫描 token
	switch first, firstwidth := utf8.DecodeRune(data[start:]); first {
	// 读引号内的 rune 或串，从引号后开始扫描
	case '\'', '"', '`':
		for width, i := 0, start+firstwidth; i < len(data); i += width {
			var r rune
			r, width = utf8.DecodeRune(data[i:])
			if r == '\\' { // 跳过逃逸符
				_, newwidth := utf8.DecodeRune(data[i+width:])
				width += newwidth
			} else if r == first { // 找到终点
				return i + width, data[start : i+width], nil
			}
		}
	// 读标识符、数字等，读到空格或注释符为止
	default:
		for width, i := 0, start; i < len(data); i += width {
			var r rune
			r, width = utf8.DecodeRune(data[i:])
			if unicode.IsSpace(r) || r == '%' {
				return i, data[start:i], nil
			}
		}
	}
	// 进入 EOF，剩下的部分全是一个 token（规则不足）
	if atEOF && len(data) > start {
		return len(data), data[start:], nil
	}
	// 要求更长的数据
	return 0, nil, nil
}
