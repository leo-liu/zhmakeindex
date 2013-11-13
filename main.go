package main

import (
	"flag"

//	"fmt"
)

func main() {
	option := NewOptions()
	option.parse()
	if !option.valid() {
		return
	}

	instyle, outstyle := NewStyles(option)

	for i, _ := range option.input {
		in := NewInputIndex(i, option, instyle)
		out := NewOutputIndex(in, option, outstyle)
		out.Output()
	}

}

type Options struct {
	comp          bool
	stdin         bool
	output        string
	page          string
	quiet         bool
	disable_range bool
	style         string
	log           string
	input         []string
}

func NewOptions() *Options {
	o := new(Options)
	flag.BoolVar(&o.comp, "c", false, "忽略条目首尾空格")
	flag.BoolVar(&o.stdin, "i", false, "从标准输入读取")
	flag.StringVar(&o.output, "o", "", "输出文件")
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
