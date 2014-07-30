// $Id$

// +build ignore

package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"unicode"
)

func init() {
	log.SetFlags(log.Lshortfile)
}

func main() {
	outdir := flag.String("d", ".", "输出目录")
	output_stroke := flag.Bool("stroke", true, "输出笔顺表")
	output_reading := flag.Bool("reading", true, "输出读音表")
	output_radical := flag.Bool("radical", true, "输出部首表")
	flag.Parse()
	if *output_stroke {
		make_stroke_table(*outdir)
	}
	if *output_reading {
		make_reading_table(*outdir)
	}
	if *output_radical {
		make_radical_table(*outdir)
	}
}

const MAX_CODEPOINT = 0x40000 // 覆盖 Unicode 第 0、1、2、3 平面

func make_stroke_table(outdir string) {
	var CJKstrokes [MAX_CODEPOINT][]byte
	var maxStroke int = 0
	var unicodeVersion string
	// 使用海峰五笔码表数据，生成笔顺表
	sunwb_file, err := os.Open("sunwb_strokeorder.txt")
	if err != nil {
		log.Fatalln(err)
	}
	defer sunwb_file.Close()
	scanner := bufio.NewScanner(sunwb_file)
	for i := 1; scanner.Scan(); i++ {
		if scanner.Err() != nil {
			log.Fatalln(scanner.Err())
		}
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) != 2 ||
			len([]rune(fields[0])) != 1 ||
			strings.IndexFunc(fields[1], isNotDigit) != -1 {
			log.Printf("笔顺文件第 %d 行语法错误，忽略。\n", i)
			continue
		}
		var r rune = []rune(fields[0])[0]
		var order []byte
		for _, rdigit := range fields[1] {
			digit, _ := strconv.ParseInt(string(rdigit), 10, 8)
			order = append(order, byte(digit))
		}
		CJKstrokes[r] = order
		if len(order) > maxStroke {
			maxStroke = len(order)
		}
	}
	// 使用 Unihan 数据库，读取笔画数补全其他字符
	unihan_file, err := os.Open("Unihan_DictionaryLikeData.txt")
	if err != nil {
		log.Fatalln(err)
	}
	defer unihan_file.Close()
	scanner = bufio.NewScanner(unihan_file)
	for scanner.Scan() {
		if scanner.Err() != nil {
			log.Fatalln(scanner.Err())
		}
		line := scanner.Text()
		if strings.Contains(line, "Unicode version:") {
			unicodeVersion = strings.TrimPrefix(line, "# ")
		}
		if strings.HasPrefix(line, "U+") && strings.Contains(line, "kTotalStrokes") {
			fields := strings.Split(line, "\t")
			var r rune
			fmt.Sscanf(fields[0], "U+%X", &r)
			var stroke int
			fmt.Sscanf(fields[2], "%d", &stroke)
			if CJKstrokes[r] != nil { // 笔顺数据已有，检查一致性
				if stroke != len(CJKstrokes[r]) {
					log.Printf("U+%04X (%c) 的笔顺数据（%d 画）与 unihan 笔画数（%d 画）不一致，跳过 unihan 数据\n",
						r, r, len(CJKstrokes[r]), stroke)
				}
			} else { // 无笔顺数据，假定每个笔画都是 6 号（未知）
				var order = make([]byte, stroke)
				for i := range order {
					order[i] = 6
				}
				CJKstrokes[r] = order
				if stroke > maxStroke {
					maxStroke = stroke
				}
			}
		}
	}
	// 输出笔顺表
	outfile, err := os.Create(path.Join(outdir, "strokes.go"))
	if err != nil {
		log.Fatalln(err)
	}
	defer outfile.Close()
	fmt.Fprintln(outfile, `// 这是由程序自动生成的文件，请不要直接编辑此文件`)
	fmt.Fprintln(outfile, `// 笔顺来源：sunwb_strokeorder.txt`)
	fmt.Fprintln(outfile, `// 笔画数来源：Unihan_DictionaryLikeData.txt`)
	fmt.Fprintf(outfile, "// Unicode 版本：%s\n", unicodeVersion)
	fmt.Fprintln(outfile, `package CJK`)
	fmt.Fprintln(outfile, `var Strokes = map[rune]string{`)
	for r, order := range CJKstrokes {
		if order == nil {
			continue
		}
		fmt.Fprintf(outfile, "\t%#x: \"", r)
		for _, s := range order {
			fmt.Fprintf(outfile, "\\x%02x", s)
		}
		fmt.Fprintf(outfile, "\", // %c\n", r)
	}
	fmt.Fprintln(outfile, `}`)
	fmt.Fprintf(outfile, "\nconst MAX_STROKE = %d\n", maxStroke)
}

func isNotDigit(r rune) bool {
	return !unicode.IsDigit(r)
}

func make_reading_table(outdir string) {
	// 读取 Unihan 读音表
	reading_table := make(map[rune]*ReadingEntry)
	reading_file, err := os.Open("Unihan_Readings.txt")
	if err != nil {
		log.Fatalln(err)
	}
	defer reading_file.Close()
	scanner := bufio.NewScanner(reading_file)
	largest := rune(0)
	var version string
	for scanner.Scan() {
		if scanner.Err() != nil {
			log.Fatalln(scanner.Err())
		}
		line := scanner.Text()
		if strings.Contains(line, "Unicode version:") {
			version = strings.TrimPrefix(line, "# ")
		}
		if strings.HasPrefix(line, "U+") {
			fields := strings.Split(line, "\t")
			var r rune
			fmt.Sscanf(fields[0], "U+%X", &r)
			if reading_table[r] == nil {
				reading_table[r] = &ReadingEntry{}
			}
			switch fields[1] {
			case "kHanyuPinyin":
				reading_table[r].HanyuPinyin = fields[2]
			case "kMandarin":
				reading_table[r].Mandarin = fields[2]
			case "kXHC1983":
				reading_table[r].XHC1983 = fields[2]
			}
			if r > largest {
				largest = r
			}
		}
	}
	// 整理所有汉字的拼音表
	out_reading_table := make([]string, largest+1)
	for k, v := range reading_table {
		pinyin := v.regular()
		numbered := NumberedPinyin(pinyin)
		out_reading_table[k] = numbered
	}
	// 单独增加数字“〇”的读音
	if out_reading_table['〇'] == "" {
		out_reading_table['〇'] = "ling2"
	}
	// 输出
	outfile, err := os.Create(path.Join(outdir, "readings.go"))
	if err != nil {
		log.Fatalln(err)
	}
	defer outfile.Close()
	fmt.Fprintln(outfile, `// 这是由程序自动生成的文件，请不要直接编辑此文件`)
	fmt.Fprintln(outfile, `// 来源：Unihan_Readings.txt`)
	fmt.Fprintln(outfile, `//`, version)
	fmt.Fprintln(outfile, `package CJK`)
	fmt.Fprintln(outfile, `var Readings = map[rune]string{`)
	for k, v := range out_reading_table {
		if v != "" {
			fmt.Fprintf(outfile, "\t%#x: %s, // %c\n", k, strconv.Quote(v), k)
		}
	}
	fmt.Fprintln(outfile, `}`)
}

type ReadingEntry struct {
	HanyuPinyin string
	Mandarin    string
	XHC1983     string
}

// 取出最常用的一个拼音
// 按如下优先次序：Mandarin -> XHC1983 -> HanyuPinyin
func (entry *ReadingEntry) regular() string {
	// kMandarin Syntax: [a-z\x{300}-\x{302}\x{304}\x{308}\x{30C}]+
	// 如 lüè
	if entry.Mandarin != "" {
		// 目前文件中没有多值情况，不过按 UAX #38 允许多值
		return strings.Split(entry.Mandarin, " ")[0]
	}
	// kXHC1983 Syntax: [0-9]{4}\.[0-9]{3}\*?(,[0-9]{4}\.[0-9]{3}\*?)*:[a-z\x{300}\x{301}\x{304}\x{308}\x{30C}]+
	// 如 1327.041:yán 1333.051:yàn
	if entry.XHC1983 != "" {
		// 第一项中第一个引号后的部分
		b := strings.Index(entry.XHC1983, ":")
		e := strings.Index(entry.XHC1983, " ")
		if e > 0 {
			return entry.XHC1983[b+1 : e]
		} else {
			return entry.XHC1983[b+1:]
		}
	}
	// kHanyuPinyin Syntax: (\d{5}\.\d{2}0,)*\d{5}\.\d{2}0:([a-z\x{300}-\x{302}\x{304}\x{308}\x{30C}]+,)*[a-z\x{300}-\x{302}\x{304}\x{308}\x{30C}]+
	// 如 10093.130:xī,lǔ 74609.020:lǔ,xī
	if entry.HanyuPinyin != "" {
		// 第一个冒号后，逗号/空格或词尾前的部分
		b := strings.Index(entry.HanyuPinyin, ":")
		e := strings.IndexAny(entry.HanyuPinyin[b:], " ,")
		if e > 0 {
			return entry.HanyuPinyin[b+1 : b+e]
		} else {
			return entry.HanyuPinyin[b+1:]
		}
	}
	// 没有汉语读音
	return ""
}

// 把拼音转换为无声调的拼音加数字声调
// 其中 ü 变为 v，轻声调号为 5，如 lǎo 转换为 lao3，lǘ 转换为 lv2
func NumberedPinyin(pinyin string) string {
	if pinyin == "" {
		return ""
	}
	numbered := []rune{}
	tone := 5
	for _, r := range pinyin {
		if Vowel[r] == 0 {
			numbered = append(numbered, r)
		} else {
			numbered = append(numbered, Vowel[r])
		}
		if Tones[r] != 0 {
			tone = Tones[r]
		}
	}
	numbered = append(numbered, []rune(strconv.Itoa(tone))...)
	return string(numbered)
}

var Vowel = map[rune]rune{
	'ā': 'a', 'á': 'a', 'ǎ': 'a', 'à': 'a',
	'ō': 'o', 'ó': 'o', 'ǒ': 'o', 'ò': 'o',
	'ē': 'e', 'é': 'e', 'ě': 'e', 'è': 'e',
	'ī': 'i', 'í': 'i', 'ǐ': 'i', 'ì': 'i',
	'ū': 'u', 'ú': 'u', 'ǔ': 'u', 'ù': 'u',
	'ǖ': 'v', 'ǘ': 'v', 'ǚ': 'v', 'ǜ': 'v', 'ü': 'v',
	'ń': 'n', 'ň': 'n', 'ǹ': 'n',
}

var Tones = map[rune]int{
	'ā': 1, 'ō': 1, 'ē': 1, 'ī': 1, 'ū': 1, 'ǖ': 1,
	'á': 2, 'ó': 2, 'é': 2, 'í': 2, 'ú': 2, 'ǘ': 2, 'ń': 2,
	'ǎ': 3, 'ǒ': 3, 'ě': 3, 'ǐ': 3, 'ǔ': 3, 'ǚ': 3, 'ň': 3,
	'à': 4, 'ò': 4, 'è': 4, 'ì': 4, 'ù': 4, 'ǜ': 4, 'ǹ': 4,
	'ü': 5,
}

func make_radical_table(outdir string) {
	// 读入部首
	CJKRadical := read_radicals()
	// 读入部首、除部首笔画
	version, CJKRadicalStrokes := read_radical_strokes()
	// 单独增加数字“〇”的部首、除部首笔画（乙部 0 画）
	CJKRadicalStrokes['〇'] = MakeRadicalStroke('〇', 5, 0)
	// 输出
	outfile, err := os.Create(path.Join(outdir, "radicalstrokes.go"))
	if err != nil {
		log.Fatalln(err)
	}
	defer outfile.Close()
	fmt.Fprintln(outfile, `// 这是由程序自动生成的文件，请不要直接编辑此文件
// 部首来源：CJKRadicals.txt
// 部首笔画数来源：Unihan_RadicalStrokeCounts.txt`)
	fmt.Fprintln(outfile, `//`, version)
	fmt.Fprintln(outfile, `package CJK

// 康熙字典部首
// 未包括 U+2F00 至 U+2FD5 等部首专用符号
type Radical struct {
	Origin     rune // 原部首的对应汉字
	Simplified rune // 简化部首
}

const MAX_RADICAL = 214

var Radicals = [MAX_RADICAL + 1]Radical{`)
	for i := 1; i < MAX_RADICAL+1; i++ {
		fmt.Fprintf(outfile, "\t%d: {%#x, %#x}, // %c",
			i, CJKRadical[i].Origin, CJKRadical[i].Simplified, CJKRadical[i].Origin)
		if CJKRadical[i].Simplified == 0 {
			fmt.Fprintln(outfile)
		} else {
			fmt.Fprintf(outfile, " (%c)\n", CJKRadical[i].Simplified)
		}
	}
	fmt.Fprintln(outfile, "}\n")
	fmt.Fprintln(outfile, `// 部首与除部首笔画数
// 前两个字节分别放部首和除部首笔画数，后面放字符本身的 UTF-8 编码，可直接排序
type RadicalStroke string

func (rs RadicalStroke) Radical() int {
	return int(rs[0])
}

func (rs RadicalStroke) Stroke() int {
	return int(rs[1])
}

var RadicalStrokes = map[rune]RadicalStroke{`)
	for r, rs := range CJKRadicalStrokes {
		if rs != "" {
			fmt.Fprintf(outfile, "\t%#x: %+q, // %c\n", r, rs, r)
		}
	}
	fmt.Fprintln(outfile, "}")
}

// 康熙字典部首
// 未包括 U+2F00 至 U+2FD5 等部首专用符号
type Radical struct {
	Origin     rune // 原部首的对应汉字
	Simplified rune // 简化部首
}

const MAX_RADICAL = 214

// 读取 CJKRadicals.txt 获取康熙字典部首表
func read_radicals() [MAX_RADICAL + 1]Radical {
	radical_file, err := os.Open("CJKRadicals.txt")
	if err != nil {
		log.Fatalln(err)
	}
	defer radical_file.Close()
	var CJKRadical [MAX_RADICAL + 1]Radical
	scanner := bufio.NewScanner(radical_file)
	for scanner.Scan() {
		if scanner.Err() != nil {
			log.Fatalln(scanner.Err())
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Split(line, ";")
		indexstr := fields[0]
		var index int
		fmt.Sscanf(fields[0], "%d", &index)
		var char rune
		fmt.Sscanf(fields[2], "%X", &char)
		if strings.HasSuffix(indexstr, "'") {
			CJKRadical[index].Simplified = char
		} else {
			CJKRadical[index].Origin = char
		}
	}
	return CJKRadical
}

// 部首与除部首笔画数
// 前两个字节分别放部首和除部首笔画数，后面放字符本身的 UTF-8 编码，可直接排序
type RadicalStroke string

func (rs RadicalStroke) Radical() int {
	return int(rs[0])
}

func (rs RadicalStroke) Stroke() int {
	return int(rs[1])
}

func MakeRadicalStroke(r rune, radical, stroke int) RadicalStroke {
	buf := []byte{byte(radical), byte(stroke)}
	return RadicalStroke(buf) + RadicalStroke(r)
}

// 读取 Unihan_IRGSources.txt 获取部首笔画数表
func read_radical_strokes() (version string, CJKRadicalStrokes []RadicalStroke) {
	radical_file, err := os.Open("Unihan_IRGSources.txt")
	if err != nil {
		log.Fatalln(err)
	}
	defer radical_file.Close()
	CJKRadicalStrokes = make([]RadicalStroke, MAX_CODEPOINT)
	scanner := bufio.NewScanner(radical_file)
	for scanner.Scan() {
		if scanner.Err() != nil {
			log.Fatalln(scanner.Err())
		}
		line := strings.TrimSpace(scanner.Text())
		if strings.Contains(line, "Unicode version:") {
			version = strings.TrimPrefix(line, "# ")
		}
		if strings.HasPrefix(line, "U+") {
			fields := strings.Split(line, "\t")
			if fields[1] == "kRSUnicode" {
				var r rune
				fmt.Sscanf(fields[0], "U+%X", &r)
				var radical, stroke int
				if strings.ContainsRune(fields[2], '\'') {
					fmt.Sscanf(fields[2], "%d'.%d", &radical, &stroke)
				} else {
					fmt.Sscanf(fields[2], "%d.%d", &radical, &stroke)
				}
				CJKRadicalStrokes[r] = MakeRadicalStroke(r, radical, stroke)
			}
		}
	}
	return
}
