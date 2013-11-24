package main

import (
	"flag"
	"io"
	"log"
	"os"
	"path/filepath"
)

var debug = log.New(os.Stderr, "DEBUG: ", log.Lshortfile)

func init() {
	log.SetFlags(0)
	log.SetPrefix("zhmakeindex: ")
}

func main() {
	option := NewOptions()
	option.parse()

	setupLog(option.log)

	instyle, outstyle := NewStyles(option.style)

	in := NewInputIndex(&option.InputOptions, instyle)
	out := NewOutputIndex(in, &option.OutputOptions, outstyle)
	out.Output()
}

type Options struct {
	InputOptions
	OutputOptions
	style string
	log   string
}

type InputOptions struct {
	compress bool
	stdin    bool
	input    []string
}

type OutputOptions struct {
	output        string
	sort          string
	page          string
	quiet         bool
	disable_range bool
}

func NewOptions() *Options {
	o := new(Options)
	flag.BoolVar(&o.compress, "c", false, "忽略条目首尾空格")
	flag.BoolVar(&o.stdin, "i", false, "从标准输入读取")
	flag.StringVar(&o.output, "o", "", "输出文件")
	flag.StringVar(&o.sort, "x", "pinyin", "中文排序方式，可以使用 pinyin 或 stroke")
	// flag.StringVar(&o.page, "p", "", "设置起始页码") // 未实现
	flag.BoolVar(&o.quiet, "q", false, "静默模式，不输出错误信息")
	flag.BoolVar(&o.disable_range, "r", false, "禁用自动生成页码区间")
	flag.StringVar(&o.style, "s", "", "格式文件名")
	flag.StringVar(&o.log, "t", "", "日志文件名")
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
}

func setupLog(logname string) {
	if logname == "" {
		// 只使用标准错误流
		return
	}
	flog, err := os.Create(logname)
	if err != nil {
		log.Fatalln(err)
	}
	log.SetOutput(io.MultiWriter(os.Stderr, flog))
}

// 删除文件后缀名
func stripExt(fpath string) string {
	ext := filepath.Ext(fpath)
	return fpath[:len(fpath)-len(ext)]
}
