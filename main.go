package main

import (
	"flag"
	"log"
	"os"
)

var debug = log.New(os.Stderr, "DEBUG: ", log.Lshortfile)

func init() {
	log.SetFlags(log.Lshortfile)
}

func main() {
	option := NewOptions()
	option.parse()
	if !option.valid() {
		return
	}

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
	flag.StringVar(&o.page, "p", "", "页码设置")
	flag.BoolVar(&o.quiet, "q", false, "静默模式，不输出错误信息")
	flag.BoolVar(&o.disable_range, "r", false, "禁用自动生成页码区间")
	flag.StringVar(&o.style, "s", "", "格式文件名")
	flag.StringVar(&o.log, "t", "", "日志文件名")
	return o
}

func (o *Options) parse() {
	flag.Parse()
	o.input = flag.Args()
}

func (o *Options) valid() bool {
	return true
}
