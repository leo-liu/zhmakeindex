package main

import (
	"bufio"
	"io"
)

// 用于逐字符读取的辅助 Reader，提供行号支持
type NumberdReader struct {
	reader   *bufio.Reader
	line     int
	lastRune rune
}

func NewNumberdReader(reader io.Reader) *NumberdReader {
	return &NumberdReader{
		reader:   bufio.NewReader(reader),
		line:     1,
		lastRune: -1,
	}
}

func (reader *NumberdReader) ReadRune() (r rune, size int, err error) {
	r, size, err = reader.reader.ReadRune()
	reader.lastRune = r
	if r == '\n' {
		reader.line++
	}
	return r, size, err
}

func (reader *NumberdReader) UnreadRune() error {
	err := reader.reader.UnreadRune()
	if err == nil {
		if reader.lastRune == '\n' {
			reader.line--
		}
		reader.lastRune = -1
	}
	return err
}

func (reader *NumberdReader) SkipLine() error {
	reader.line++
	reader.lastRune = -1
	_, err := reader.reader.ReadBytes('\n')
	return err
}

func (reader *NumberdReader) Line() int {
	return reader.line
}
