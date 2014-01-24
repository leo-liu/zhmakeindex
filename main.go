// $Id: main.go,v f556a652ef7a 2014/01/24 14:58:07 LeoLiu $

// zhmakeindex: 带中文支持的 makeindex 实现
package main

import (
	"flag"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"code.google.com/p/go.text/encoding"
	"code.google.com/p/go.text/encoding/simplifiedchinese"
	"code.google.com/p/go.text/encoding/traditionalchinese"
	"code.google.com/p/go.text/encoding/unicode"
	"code.google.com/p/go.text/transform"
)

var (
	ProgramAuthor   = "刘海洋"
	ProgramVersion  = "beta"
	ProgramRevision = stripDollors("$Revision: f556a652ef7a $", "Revision:")
)

var debug = log.New(os.Stderr, "DEBUG: ", log.Lshortfile)

func init() {
	log.SetFlags(0)
	log.SetPrefix("")
}

func main() {
	option := NewOptions()
	option.parse()

	setupLog(option)

	log.Printf("zhmakeindex 版本：%s (r%s)\t作者：%s\n", ProgramVersion, ProgramRevision, ProgramAuthor)

	if option.style != "" {
		log.Println("正在读取格式文件……")
	}
	instyle, outstyle := NewStyles(&option.StyleOptions)

	in := NewInputIndex(&option.InputOptions, instyle)
	log.Printf("合并后共 %d 项。\n", len(*in))

	log.Println("正在排序……")
	out := NewOutputIndex(in, &option.OutputOptions, outstyle)

	log.Println("正在输出……")
	out.Output(&option.OutputOptions)

	if option.output != "" {
		log.Printf("输出文件写入 %s\n", option.output)
	}
	if option.log != "" {
		log.Printf("日志文件写入 %s\n", option.log)
	}
}

type Options struct {
	InputOptions
	OutputOptions
	StyleOptions
	encoding       string
	style_encoding string
	log            string
	quiet          bool
}

type InputOptions struct {
	compress bool
	stdin    bool
	decoder  transform.Transformer // 由 encoding 生成
	input    []string
}

type OutputOptions struct {
	encoder       transform.Transformer // 由 encoding 生成
	output        string
	sort          string
	page          string
	strict        bool
	disable_range bool
}

type StyleOptions struct {
	style         string
	style_decoder transform.Transformer // 由 style_encoding 生成
}

func NewOptions() *Options {
	o := new(Options)
	// flag.BoolVar(&o.compress, "c", false, "忽略条目首尾空格") // 未实现
	flag.BoolVar(&o.stdin, "i", false, "从标准输入读取")
	flag.StringVar(&o.output, "o", "", "输出文件")
	flag.StringVar(&o.sort, "z", "pinyin",
		"中文排序方式，可以使用 pinyin (reading)、bihua (stroke) 或 bushou (radical)")
	// flag.StringVar(&o.page, "p", "", "设置起始页码") // 未实现
	flag.BoolVar(&o.quiet, "q", false, "静默模式，不输出错误信息")
	flag.BoolVar(&o.disable_range, "r", false, "禁用自动生成页码区间")
	flag.BoolVar(&o.strict, "strict", false, "严格区分不同 encapsulated 命令的页码")
	flag.StringVar(&o.style, "s", "", "格式文件名")
	flag.StringVar(&o.log, "t", "", "日志文件名")
	flag.StringVar(&o.encoding, "enc", "utf-8", "读写索引文件的编码")
	flag.StringVar(&o.style_encoding, "senc", "utf-8", "格式文件的编码")
	return o
}

func (o *Options) parse() {
	flag.Parse()

	o.input = flag.Args()
	// 整理输入文件，没有后缀时，加上默认后缀 .idx
	for i := range o.input {
		o.input[i] = filepath.Clean(o.input[i])
		if filepath.Ext(o.input[i]) == "" {
			o.input[i] = o.input[i] + ".idx"
		}
	}
	// 错误的参数组合
	if len(o.input) > 0 && o.stdin {
		log.Fatalln("不能同时从文件和标准输入流读取输入")
	} else if len(o.input) == 0 && !o.stdin {
		log.Fatalln("没有输入文件")
	}
	// 不指定输出文件且不使用标准输入时，使用第一个输入文件的主文件名 + ".ind" 后缀
	if o.output == "" && !o.stdin {
		o.output = stripExt(o.input[0]) + ".ind"
	}
	// 不指定输入文件且不使用标准输入时，使用第一个输入文件的主文件名 + ".ilg" 后缀
	if o.log == "" && !o.stdin {
		o.log = stripExt(o.input[0]) + ".ilg"
	}

	// 检查并设置 IO 编码
	encoding := checkEncoding(o.encoding)
	o.encoder = encoding.NewEncoder()
	o.decoder = encoding.NewDecoder()

	// 检查并设置格式文件编码
	styleEncoding := checkEncoding(o.style_encoding)
	o.style_decoder = styleEncoding.NewDecoder()
}

// 检查编码名，返回编码
func checkEncoding(encodingName string) encoding.Encoding {
	var encodingMap = map[string]encoding.Encoding{
		"utf-8":   encoding.Nop,
		"utf-16":  unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM),
		"gb18030": simplifiedchinese.GB18030,
		"gbk":     simplifiedchinese.GBK,
		"big5":    traditionalchinese.Big5,
	}
	encoding, ok := encodingMap[strings.ToLower(encodingName)]
	if !ok {
		log.Printf("编码 '%s' 是无效编码\n", encodingName)
		var supported string
		for enc := range encodingMap {
			supported += enc + " "
		}
		log.Fatalf("支持的编码有（不区分大小写）：%s\n", supported)
	}
	return encoding
}

func setupLog(option *Options) {
	var stderr io.Writer = os.Stderr
	if option.quiet {
		stderr = ioutil.Discard
	}
	if option.log == "" {
		// 只使用标准错误流
		return
	}
	flog, err := os.Create(option.log)
	if err != nil {
		log.Fatalln(err)
	}
	log.SetOutput(io.MultiWriter(stderr, flog))
}

// 删除文件后缀名
func stripExt(fpath string) string {
	ext := filepath.Ext(fpath)
	return fpath[:len(fpath)-len(ext)]
}

func stripDollors(svnstring, prefix string) string {
	if prefix == "" {
		prefix = "$"
	}
	begin := strings.Index(svnstring, prefix) + len(prefix)
	end := strings.LastIndex(svnstring, "$")
	return strings.TrimSpace(svnstring[begin:end])
}
