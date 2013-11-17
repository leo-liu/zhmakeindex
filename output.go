package main

import (
	"fmt"
	"log"
	"os"
)

// 分类排序方式
type IndexSorter interface {
	SortIndex(input *InputIndex) *OutputIndex
}

// 输出索引
type OutputIndex struct {
	groups []IndexGroup
	style  *OutputStyle
	option *OutputOptions
}

func NewOutputIndex(input *InputIndex, option *OutputOptions, style *OutputStyle) *OutputIndex {
	var sorter IndexSorter
	switch option.sort {
	case "stroke":
		sorter = &StrokesSorter{}
	default:
		log.Fatalln("未知排序方式")
	}
	outindex := sorter.SortIndex(input)
	outindex.style = style
	outindex.option = option
	return outindex
}

// 按格式输出索引项
func (o *OutputIndex) Output() {
	outfile, err := os.Create(o.option.output)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Fprint(outfile, o.style.preamble)
	fmt.Fprint(outfile, o.style.postamble)
	defer outfile.Close()
}

// 一个输出项目组
type IndexGroup struct {
	name  string
	items []IndexItem
}

// 一个输出项，包括级别、文字、用来排序的键、一系列页码区间，下一级的项列表
type IndexItem struct {
	level     int
	text      string
	key       string
	page      []PageRange
	nextlevel []IndexItem
}

type PageRange struct {
	tag   RangeType
	begin string
	end   string
}
