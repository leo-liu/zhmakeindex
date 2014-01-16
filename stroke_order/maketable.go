// $Id: maketable.go,v 7e2fc0393fd4 2014/01/16 20:03:56 leoliu $

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
	flag.Parse()
	make_stroke_order_table(*outdir)
}

// 使用海峰五笔码表数据，生成笔顺表
func make_stroke_order_table(outdir string) {
	reading_file, err := os.Open("sunwb_strokeorder.txt")
	if err != nil {
		log.Fatalln(err)
	}
	defer reading_file.Close()
	outfile, err := os.Create(path.Join(outdir, "stroke_order.go"))
	if err != nil {
		log.Fatalln(err)
	}
	defer outfile.Close()

	fmt.Fprintln(outfile, `// 这是由程序自动生成的文件，请不要直接编辑此文件`)
	fmt.Fprintln(outfile, `// 来源：sunwb_strokeorder.txt`)
	fmt.Fprintln(outfile, `package main`)
	fmt.Fprintln(outfile, `var CJKstrokeOrder = map[rune] []int8{`)

	scanner := bufio.NewScanner(reading_file)
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
			log.Printf("第 %d 行语法错误，忽略。\n", i)
			continue
		}
		var r rune = []rune(fields[0])[0]
		var order []int8
		for _, rdigit := range fields[1] {
			digit, _ := strconv.ParseInt(string(rdigit), 10, 8)
			order = append(order, int8(digit))
		}
		fmt.Fprintf(outfile, "\t%#x: {", r)
		for _, stroke := range order {
			fmt.Fprintf(outfile, "%d,", stroke)
		}
		fmt.Fprintf(outfile, "}, // %c\n", r)
	}

	fmt.Fprintln(outfile, `}`)
}

func isNotDigit(r rune) bool {
	return !unicode.IsDigit(r)
}
