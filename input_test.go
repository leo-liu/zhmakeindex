package main

import (
	"io"
	"strings"
	"testing"

	"github.com/yasushi-saito/rbtree"
)

func TestScanIndexEntry(t *testing.T) {
	style := NewInputStyle()
	option := &InputOptions{}

	tests := []struct {
		name      string
		input     string
		levels    []IndexEntryLevel
		encap     string
		rangetype RangeType
		pageNum   int
		pageFmt   NumFormat
	}{
		{
			name:      "simple entry",
			input:     `\indexentry{foo}{1}`,
			levels:    []IndexEntryLevel{{key: "foo", text: "foo"}},
			rangetype: PAGE_NORMAL,
			pageNum:   1,
			pageFmt:   NUM_ARABIC,
		},
		{
			name:  "multi-level entry",
			input: `\indexentry{A!B!C}{5}`,
			levels: []IndexEntryLevel{
				{key: "A", text: "A"},
				{key: "B", text: "B"},
				{key: "C", text: "C"},
			},
			rangetype: PAGE_NORMAL,
			pageNum:   5,
			pageFmt:   NUM_ARABIC,
		},
		{
			name:      "key@text syntax",
			input:     `\indexentry{key@text}{2}`,
			levels:    []IndexEntryLevel{{key: "key", text: "text"}},
			rangetype: PAGE_NORMAL,
			pageNum:   2,
			pageFmt:   NUM_ARABIC,
		},
		{
			name:      "encap syntax",
			input:     `\indexentry{foo|textbf}{3}`,
			levels:    []IndexEntryLevel{{key: "foo", text: "foo"}},
			encap:     "textbf",
			rangetype: PAGE_NORMAL,
			pageNum:   3,
			pageFmt:   NUM_ARABIC,
		},
		{
			name:      "range open",
			input:     `\indexentry{foo|(}{10}`,
			levels:    []IndexEntryLevel{{key: "foo", text: "foo"}},
			rangetype: PAGE_OPEN,
			pageNum:   10,
			pageFmt:   NUM_ARABIC,
		},
		{
			name:      "range close",
			input:     `\indexentry{foo|)}{15}`,
			levels:    []IndexEntryLevel{{key: "foo", text: "foo"}},
			rangetype: PAGE_CLOSE,
			pageNum:   15,
			pageFmt:   NUM_ARABIC,
		},
		{
			name:      "range open with encap",
			input:     `\indexentry{foo|(textit}{7}`,
			levels:    []IndexEntryLevel{{key: "foo", text: "foo"}},
			encap:     "textit",
			rangetype: PAGE_OPEN,
			pageNum:   7,
			pageFmt:   NUM_ARABIC,
		},
		{
			name:      "roman numeral page",
			input:     `\indexentry{bar}{xiv}`,
			levels:    []IndexEntryLevel{{key: "bar", text: "bar"}},
			rangetype: PAGE_NORMAL,
			pageNum:   14,
			pageFmt:   NUM_ROMAN_LOWER,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := NewNumberdReader(strings.NewReader(tt.input))
			entry, err := ScanIndexEntry(reader, option, style)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(entry.level) != len(tt.levels) {
				t.Fatalf("expected %d levels, got %d", len(tt.levels), len(entry.level))
			}
			for i, lv := range tt.levels {
				if entry.level[i].key != lv.key {
					t.Errorf("level[%d].key = %q, want %q", i, entry.level[i].key, lv.key)
				}
				if entry.level[i].text != lv.text {
					t.Errorf("level[%d].text = %q, want %q", i, entry.level[i].text, lv.text)
				}
			}
			if len(entry.pagelist) != 1 {
				t.Fatalf("expected 1 page, got %d", len(entry.pagelist))
			}
			page := entry.pagelist[0]
			if page.encap != tt.encap {
				t.Errorf("encap = %q, want %q", page.encap, tt.encap)
			}
			if page.rangetype != tt.rangetype {
				t.Errorf("rangetype = %v, want %v", page.rangetype, tt.rangetype)
			}
			if len(page.numbers) < 1 {
				t.Fatal("no page numbers")
			}
			if page.numbers[0].num != tt.pageNum {
				t.Errorf("page num = %d, want %d", page.numbers[0].num, tt.pageNum)
			}
			if page.numbers[0].format != tt.pageFmt {
				t.Errorf("page format = %d, want %d", page.numbers[0].format, tt.pageFmt)
			}
		})
	}
}

func TestScanIndexEntryEOF(t *testing.T) {
	style := NewInputStyle()
	option := &InputOptions{}
	reader := NewNumberdReader(strings.NewReader(""))
	_, err := ScanIndexEntry(reader, option, style)
	if err != io.EOF {
		t.Errorf("expected io.EOF, got %v", err)
	}
}

func TestScanIndexEntrySyntaxError(t *testing.T) {
	style := NewInputStyle()
	option := &InputOptions{}
	reader := NewNumberdReader(strings.NewReader("not an index entry"))
	_, err := ScanIndexEntry(reader, option, style)
	if err != ScanSyntaxError {
		t.Errorf("expected ScanSyntaxError, got %v", err)
	}
}

func TestCompareIndexEntry(t *testing.T) {
	tests := []struct {
		name string
		a, b *IndexEntry
		want int
	}{
		{
			name: "equal entries",
			a:    &IndexEntry{level: []IndexEntryLevel{{key: "foo", text: "foo"}}},
			b:    &IndexEntry{level: []IndexEntryLevel{{key: "foo", text: "foo"}}},
			want: 0,
		},
		{
			name: "a less than b by key",
			a:    &IndexEntry{level: []IndexEntryLevel{{key: "abc", text: "abc"}}},
			b:    &IndexEntry{level: []IndexEntryLevel{{key: "def", text: "def"}}},
			want: -1,
		},
		{
			name: "parent less than child",
			a:    &IndexEntry{level: []IndexEntryLevel{{key: "A", text: "A"}}},
			b:    &IndexEntry{level: []IndexEntryLevel{{key: "A", text: "A"}, {key: "B", text: "B"}}},
			want: -1,
		},
		{
			name: "child greater than parent",
			a:    &IndexEntry{level: []IndexEntryLevel{{key: "A", text: "A"}, {key: "B", text: "B"}}},
			b:    &IndexEntry{level: []IndexEntryLevel{{key: "A", text: "A"}}},
			want: 1,
		},
		{
			name: "same key different text",
			a:    &IndexEntry{level: []IndexEntryLevel{{key: "foo", text: "aaa"}}},
			b:    &IndexEntry{level: []IndexEntryLevel{{key: "foo", text: "zzz"}}},
			want: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CompareIndexEntry(rbtree.Item(tt.a), rbtree.Item(tt.b))
			if (tt.want < 0 && got >= 0) || (tt.want > 0 && got <= 0) || (tt.want == 0 && got != 0) {
				t.Errorf("CompareIndexEntry() = %d, want sign of %d", got, tt.want)
			}
		})
	}
}

func TestSkipspaces(t *testing.T) {
	style := NewInputStyle()

	tests := []struct {
		name     string
		input    string
		wantRune rune
	}{
		{
			name:     "skip whitespace",
			input:    "  \n  foo",
			wantRune: 'f',
		},
		{
			name:     "skip comment",
			input:    "%comment\nfoo",
			wantRune: 'f',
		},
		{
			name:     "no whitespace",
			input:    "bar",
			wantRune: 'b',
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := NewNumberdReader(strings.NewReader(tt.input))
			err := skipspaces(reader, style)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			r, _, err := reader.ReadRune()
			if err != nil {
				t.Fatalf("ReadRune error: %v", err)
			}
			if r != tt.wantRune {
				t.Errorf("next rune = %c, want %c", r, tt.wantRune)
			}
		})
	}
}
