package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

func init() {
	log.SetFlags(log.Lshortfile)
}

func main() {
	make_reading_table()
}

func make_reading_table() {
	// 读取 Unihan 读音表
	reading_table := make(map[rune]*ReadingEntry)
	reading_file, err := os.Open("Unihan_Readings.txt")
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer reading_file.Close()
	scanner := bufio.NewScanner(reading_file)
	largest := rune(0)
	for scanner.Scan() {
		if scanner.Err() != nil {
			log.Fatalln(scanner.Err().Error())
		}
		line := scanner.Text()
		if strings.HasPrefix(line, "U+") {
			fields := strings.Split(line, "\t")
			var r rune
			fmt.Sscanf(fields[0], "U+%X", &r)
			if reading_table[r] == nil {
				reading_table[r] = &ReadingEntry{}
			}
			switch fields[1] {
			case "kHanyuPinlu":
				reading_table[r].HanyuPinlu = fields[2]
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
	// 输出
	for _, v := range out_reading_table {
		fmt.Printf("%s,\n", strconv.Quote(v))
	}
}

type ReadingEntry struct {
	HanyuPinlu  string
	HanyuPinyin string
	Mandarin    string
	XHC1983     string
}

// 取出最常用的一个拼音
// 按如下优先次序：HanyuPinlu -> Mandarin -> XHC1983 -> HanyuPinyin
func (entry *ReadingEntry) regular() string {
	// xHanyuPinlu Syntax: [a-z\x{300}-\x{302}\x{304}\x{308}\x{30C}]+\([0-9]+\)
	// 如 cān(525) shēn(25)
	if entry.HanyuPinlu != "" {
		// 第一个括号之前的部分即可
		return strings.Split(entry.HanyuPinlu, "(")[0]
	}
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
}

var Tones = map[rune]int{
	'ā': 1, 'ō': 1, 'ē': 1, 'ī': 1, 'ū': 1, 'ǖ': 1,
	'á': 2, 'ó': 2, 'é': 2, 'í': 2, 'ú': 2, 'ǘ': 2,
	'ǎ': 3, 'ǒ': 3, 'ě': 3, 'ǐ': 3, 'ǔ': 3, 'ǚ': 3,
	'à': 4, 'ò': 4, 'è': 4, 'ì': 4, 'ù': 4, 'ǜ': 4,
	'ü': 5,
}
