// $Id$

package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"golang.org/x/text/transform"
)

// 输出索引
type OutputIndex struct {
	groups []IndexGroup
	style  *OutputStyle
	option *OutputOptions
}

func NewOutputIndex(input *InputIndex, option *OutputOptions, style *OutputStyle) *OutputIndex {
	sorter := NewIndexSorter(option.sort)
	outindex := sorter.SortIndex(input, style, option)
	outindex.style = style
	outindex.option = option
	return outindex
}

// 按格式输出索引项
// suffix_2p, suffix_3p, suffix_mp 暂未实现
// line_max, indent_space, indent_length 未实现
func (o *OutputIndex) Output(option *OutputOptions) {
	var writer io.WriteCloser
	if o.option.output == "" {
		writer = os.Stdout
	} else {
		var err error
		writer, err = os.Create(o.option.output)
		if err != nil {
			log.Fatalln(err)
		}
		defer writer.Close()
	}
	writer = transform.NewWriter(writer, option.encoder)

	fmt.Fprint(writer, o.style.preamble)
	first_group := true
	for _, group := range o.groups {
		if group.items == nil {
			continue
		}
		if first_group {
			first_group = false
		} else {
			fmt.Fprint(writer, o.style.group_skip)
		}
		if o.style.headings_flag != 0 {
			fmt.Fprintf(writer, "%s%s%s", o.style.heading_prefix, group.name, o.style.heading_suffix)
		}
		for i, item := range group.items {
			// debug.Println(i, item)
			// 如果修改一下 OutputStyle 的数据结构，容易改成任意层的索引
			switch item.level {
			case 0:
				fmt.Fprintf(writer, "%s%s", o.style.item_0, item.text)
				writePage(writer, 0, item.page, o.style)
			case 1:
				if last := group.items[i-1]; last.level == 0 {
					if last.page != nil {
						fmt.Fprint(writer, o.style.item_01)
					} else {
						fmt.Fprint(writer, o.style.item_x1)
					}
				} else {
					fmt.Fprint(writer, o.style.item_1)
				}
				fmt.Fprint(writer, item.text)
				writePage(writer, 1, item.page, o.style)
			case 2:
				if last := group.items[i-1]; last.level == 1 {
					if last.page != nil {
						fmt.Fprint(writer, o.style.item_12)
					} else {
						fmt.Fprint(writer, o.style.item_x2)
					}
				} else {
					fmt.Fprint(writer, o.style.item_2)
				}
				fmt.Fprint(writer, item.text)
				writePage(writer, 2, item.page, o.style)
			default:
				log.Printf("索引项“%s”层次数过深，忽略此项\n", item.text)
			}
		}
	}
	fmt.Fprint(writer, o.style.postamble)
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
	if len(pageranges) != 0 {
		fmt.Fprint(out, style.delim_t)
	}
}

// 一个输出项目组
type IndexGroup struct {
	name  string
	items []IndexItem
}

// 一个输出项，包括级别、文字、一系列页码区间
type IndexItem struct {
	level int
	text  string
	page  []PageRange
}

// 用于输出的页码区间
type PageRange struct {
	begin *Page
	end   *Page
}

func (p *PageRange) Diff() int {
	return p.end.Diff(p.begin)
}

// 输出页码区间
func (p *PageRange) Write(out io.Writer, style *OutputStyle) {
	var rangestr string
	switch {
	// 单页
	case p.Diff() == 0:
		rangestr = p.begin.String()
	// 由单页合并得到的两页的区间，且未设置 suffix_2p，视为独立的两页
	case p.begin.rangetype == PAGE_NORMAL && p.end.rangetype == PAGE_NORMAL &&
		p.Diff() == 1 && style.suffix_2p == "":
		rangestr = p.begin.String() + style.delim_n + p.end.String()
	// 两页的区间，设置了 suffix_2p
	case p.Diff() == 1 && style.suffix_2p != "":
		rangestr = p.begin.String() + style.suffix_2p
	// 三页的区间，设置了 suffix_3p
	case p.Diff() == 2 && style.suffix_3p != "":
		rangestr = p.begin.String() + style.suffix_3p
	// 三页或更长的区间，设置了 suffix_mp
	case p.Diff() >= 2 && style.suffix_mp != "":
		rangestr = p.begin.String() + style.suffix_mp
	// 普通的区间
	default:
		rangestr = p.begin.String() + style.delim_r + p.end.String()
	}
	// encap 只看区间头，对不完全区间可能不总正确
	if p.begin.encap == "" {
		fmt.Fprint(out, rangestr)
	} else {
		fmt.Fprint(out, style.encap_prefix, p.begin.encap,
			style.encap_infix, rangestr, style.encap_suffix)
	}
}
