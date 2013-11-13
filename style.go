package main

func NewStyles(option *Options) (*InputStyle, *OutputStyle) {
	in := NewInputStyle()
	out := NewOutputStyle()

	return in, out
}

type InputStyle struct {
	keyword         string
	arg_open        rune
	arg_close       rune
	actual          rune
	encap           rune
	escape          rune
	level           rune
	quote           rune
	page_compositor string
	range_open      rune
	range_close     rune
}

func NewInputStyle() *InputStyle {
	in := &InputStyle{
		keyword:         "\\indexentry",
		arg_open:        '{',
		arg_close:       '}',
		actual:          '@',
		encap:           '|',
		escape:          '\\',
		level:           '!',
		quote:           '"',
		page_compositor: "-",
		range_open:      '(',
		range_close:     ')',
	}
	return in
}

type OutputStyle struct {
	preamble         string
	postamble        string
	setpage_prefix   string
	setpage_suffix   string
	group_skip       string
	headings_flag    int
	heading_prefix   string
	symhead_positive string
	symhead_negative string
	numhead_positive string
	numhead_negative string
	item_0           string
	item_1           string
	item_2           string
	item_01          string
	item_x1          string
	item_12          string
	item_x2          string
	delim_0          string
	delim_1          string
	delim_2          string
	delim_n          string
	delim_r          string
	delim_t          string
	encap_prefix     string
	encap_infix      string
	encap_suffix     string
	line_max         int
	indent_space     string
	indent_length    int
	suffix_2p        string
	suffix_3p        string
	suffix_mp        string
}

func NewOutputStyle() *OutputStyle {
	out := &OutputStyle{
		preamble:         "\\begin{theindex}\n",
		postamble:        "\n\n\\end{theindex}\n",
		setpage_prefix:   "\n  \\setcounter{page}{",
		setpage_suffix:   "}\n",
		group_skip:       "\n\n  \\indexspace\n",
		headings_flag:    0,
		heading_prefix:   "",
		symhead_positive: "Symbols",
		symhead_negative: "symbols",
		numhead_positive: "Numbers",
		numhead_negative: "numbers",
		item_0:           "\n  \\tiem ",
		item_1:           "\n    \\subitem ",
		item_2:           "\n      \\subsubitem ",
		item_01:          "\n    \\subitem ",
		item_x1:          "\n    \\subitem ",
		item_12:          "\n      \\subsubitem ",
		item_x2:          "\n      \\subsubitem ",
		delim_0:          ", ",
		delim_1:          ", ",
		delim_2:          ", ",
		delim_n:          ", ",
		delim_r:          "--",
		delim_t:          "",
		encap_prefix:     "\\",
		encap_infix:      "{",
		encap_suffix:     "}",
		line_max:         72,
		indent_space:     "\t\t",
		indent_length:    16,
		suffix_2p:        "",
		suffix_3p:        "",
		suffix_mp:        "",
	}
	return out
}
