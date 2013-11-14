package main

type InputIndex []IndexEntry

func NewInputIndex(option *Options, style *InputStyle) *InputIndex {
	in := new(InputIndex)
	return in
}

//func (ind InputIndex) Len() int {
//	return len(ind)
//}

//func (ind InputIndex) Swap(i, j int) {
//	ind[i], ind[j] = ind[j], ind[i]
//}

//func (ind InputIndex) Less(i, j int) bool {
//	return true
//}

type IndexEntry struct {
	level     []IndexEntryKV
	pagefmt   string
	page      int
	pagerange PageRange
}

type PageRange int

const (
	PAGE_NORMAL PageRange = iota
	PAGE_OPEN
	PAGE_CLOSE
)

type IndexEntryKV struct {
	key  string
	text string
}
