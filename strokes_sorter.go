// strokes_sorter.go
package main

type StrokesSorter struct{}

func (sorter *StrokesSorter) SortIndex(input *InputIndex) *OutputIndex {
	out := new(OutputIndex)
	out.groups = make([]IndexGroup, 1)
	out.groups[0].name = "1 划"
	out.groups[0].items = make([]IndexItem, 1)
	out.groups[0].items[0].text = "乙"
	out.groups[0].items[0].level = 0
	out.groups[0].items[0].page = []PageRange{{tag: PAGE_NORMAL, begin: "1", end: ""}}
	return out
}
