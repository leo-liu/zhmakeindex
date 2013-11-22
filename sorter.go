package main

import (
	"log"
	"sort"
)

// 分类排序方式
type IndexSorter interface {
	SortIndex(input *InputIndex, style *OutputStyle) *OutputIndex
}

// 页码排序器
type PageSorter struct {
	precedence map[NumFormat]int
}

func NewPageSorter(style *OutputStyle) *PageSorter {
	sorter := PageSorter{}
	sorter.precedence = make(map[NumFormat]int)
	for i, r := range style.page_precedence {
		switch r {
		case 'r':
			sorter.precedence[NUM_ROMAN_LOWER] = i
		case 'n':
			sorter.precedence[NUM_ARABIC] = i
		case 'a':
			sorter.precedence[NUM_ALPH_LOWER] = i
		case 'R':
			sorter.precedence[NUM_ROMAN_UPPER] = i
		case 'A':
			sorter.precedence[NUM_ALPH_UPPER] = i
		default:
			log.Println("page_precedence 语法错误，采用默认值")
			sorter.precedence = map[NumFormat]int{
				NUM_ROMAN_LOWER: 0,
				NUM_ARABIC:      1,
				NUM_ALPH_LOWER:  2,
				NUM_ROMAN_UPPER: 3,
				NUM_ALPH_UPPER:  4,
			}
		}
	}
	return &sorter
}

// 处理输入的页码，生成页码区间组
func (sorter *PageSorter) Sort(pages []PageInput) []PageRange {
	out := []PageRange{}
	sort.Sort(PageInputSlice{pages: pages, sorter: sorter})
	// 使用一个栈来合并页码区间
	stack := make([]PageInput, 0)
	for _, p := range pages {
		pstr := p.NumString()
		if len(stack) == 0 {
			switch p.rangetype {
			case PAGE_NORMAL:
				// 输出独立页
				out = append(out, PageRange{encap: p.encap, begin: pstr, end: pstr})
			case PAGE_OPEN:
				// 压栈
				stack = append(stack, p)
			case PAGE_CLOSE:
				log.Printf("页码区间有误，区间末尾 %s 没有匹配的区间头。\n", pstr)
				// 输出从空白到当前页的伪区间
				out = append(out, PageRange{encap: p.encap, begin: "", end: pstr})
			}
		} else {
			top := stack[len(stack)-1]
			if p.encap != top.encap || p.format != top.format {
				log.Printf("页码区间有误，未找到与 %s 匹配的区间尾。\n", stack[0].NumString())
				// 输出从当前页到空白的伪区间，并清空栈
				out = append(out, PageRange{encap: p.encap, begin: stack[0].NumString(), end: ""})
				stack = make([]PageInput, 0)
				continue
			}
			switch p.rangetype {
			case PAGE_NORMAL:
				// 什么也不做
			case PAGE_OPEN:
				// 压栈
				stack = append(stack, p)
			case PAGE_CLOSE:
				// 栈中只有一个元素时输出正常区间，弹栈
				if len(stack) == 1 {
					out = append(out, PageRange{encap: p.encap, begin: stack[0].NumString(), end: pstr})
				}
				stack = stack[:len(stack)-1]
			}
		}
	}
	return out
}

type PageInputSlice struct {
	pages  []PageInput
	sorter *PageSorter
}

func (p PageInputSlice) Len() int {
	return len(p.pages)
}

func (p PageInputSlice) Swap(i, j int) {
	p.pages[i], p.pages[j] = p.pages[j], p.pages[i]
}

// 先按 encap 类型比较，然后按页码类型，然后页码数值，最后是 rangetype，方便以后合并
func (p PageInputSlice) Less(i, j int) bool {
	a, b := p.pages[i], p.pages[j]
	if a.encap < b.encap {
		return true
	} else if a.encap > b.encap {
		return false
	}
	if p.sorter.precedence[a.format] < p.sorter.precedence[b.format] {
		return true
	} else if p.sorter.precedence[a.format] > p.sorter.precedence[b.format] {
		return false
	}
	if a.page < b.page {
		return true
	} else if a.page > b.page {
		return false
	}
	if a.rangetype < b.rangetype {
		return true
	} else {
		return false
	}
}
