package main

type indexentry struct {
	level     []indexentrykv
	pagefmt   string
	page      int
	pagerange pagerange
}

type pagerange int

const (
	page_normal pagerange = iota
	page_open   pagerange = iota
	page_close  pagerange = iota
)

type indexentrykv struct {
	sortkey string
	text    string
}
