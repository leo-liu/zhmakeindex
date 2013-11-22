package main

import (
	"sort"
)

// 分类排序方式
type IndexSorter interface {
	SortIndex(input *InputIndex, style *OutputStyle) *OutputIndex
}

// 页码排序器
type PageSorter struct{}

// 处理输入的页码，生成页码
func (sorter *PageSorter) Sort(pages []PageInput) []PageRange {
	sort.Sort(PageInputSlice{pages})
	stack := make([]PageInput, 0, len(pages))
	top := 0
	for i, p := range pages {
		if top == 0 {
		}
	}
	return nil
}

type PageInputSlice struct {
	pages []PageInput
}

func (p PageInputSlice) Len() int {
	return len(p.pages)
}

func (p PageInputSlice) Swap(i, j int) {
	p.pages[i], p.pages[j] = p.pages[j], p.pages[i]
}

// 先按 encap 类型比较，然后按页码，最后是 rangetype，方便以后合并
func (p PageInputSlice) Less(i, j int) bool {
	a, b := p.pages[i], p.pages[j]
	if a.encap < b.encap {
		return true
	} else if a.encap > b.encap {
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
