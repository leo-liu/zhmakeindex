package main

import (
	"fmt"
	"io"
	"log"
	"os"
)

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
	outindex := sorter.SortIndex(input, style)
	outindex.style = style
	outindex.option = option
	return outindex
}

// 按格式输出索引项
// suffix_2p, suffix_3p, suffix_mp 暂未实现
// line_max, indent_space, indent_length 未实现
func (o *OutputIndex) Output() {
	outfile, err := os.Create(o.option.output)
	if err != nil {
		log.Fatalln(err)
	}
	defer outfile.Close()

	fmt.Fprint(outfile, o.style.preamble)
	for _, group := range o.groups {
		if group.items == nil {
			continue
		}
		fmt.Fprint(outfile, o.style.group_skip)
		if o.style.headings_flag > 0 {
			fmt.Fprintf(outfile, "%s%s%s", o.style.heading_prefix, group.name, o.style.heading_suffix)
		} else if o.style.headings_flag < 0 {
			fmt.Fprintf(outfile, "%s%s%s", o.style.heading_prefix, group.name, o.style.heading_suffix)
		}
		for _, item0 := range group.items {
			fmt.Fprintf(outfile, "%s%s", o.style.item_0, item0.text)
			writePage(outfile, 0, item0.page, o.style)
			for i1, item1 := range item0.nextlevel {
				if i1 == 0 {
					if item0.page != nil {
						fmt.Fprint(outfile, o.style.item_01)
					} else {
						fmt.Fprint(outfile, o.style.item_x1)
					}
				} else {
					fmt.Fprint(outfile, o.style.item_1)
				}
				fmt.Fprint(outfile, item1.text)
				writePage(outfile, 1, item1.page, o.style)
				for i2, item2 := range item1.nextlevel {
					if i2 == 0 {
						if item1.page != nil {
							fmt.Fprint(outfile, o.style.item_12)
						} else {
							fmt.Fprint(outfile, o.style.item_x2)
						}
					} else {
						fmt.Fprint(outfile, o.style.item_2)
					}
					fmt.Fprint(outfile, item2.text)
					writePage(outfile, 2, item2.page, o.style)
				}
			}
		}
	}
	fmt.Fprint(outfile, o.style.postamble)
}

func writePage(out io.Writer, level int, pageranges []PageRange, style *OutputStyle) {
	if pageranges == nil {
		return
	}
	switch level {
	case 0:
		fmt.Fprint(out, style.delim_0)
	case 1:
		fmt.Fprint(out, style.delim_1)
	case 2:
		fmt.Fprint(out, style.delim_2)
	}
	for i, p := range pageranges {
		if i > 0 {
			fmt.Fprint(out, style.delim_n)
		}
		p.Write(out, style)
	}
	fmt.Fprint(out, style.delim_t)
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
	encap string
	begin string
	end   string
}

func (p *PageRange) Write(out io.Writer, style *OutputStyle) {
	if p.encap != "" {
		fmt.Fprintf(out, "%s%s%s", style.encap_prefix, p.encap, style.encap_infix)
	}
	if p.begin == p.end {
		fmt.Fprint(out, p.begin)
	} else {
		fmt.Fprintf(out, "%s%s%s", p.begin, style.delim_r, p.end)
	}
	if p.encap != "" {
		fmt.Fprint(out, style.encap_suffix)
	}
}
