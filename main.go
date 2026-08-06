package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

// 一个字符占几列宽，粗略按码位判断（宽字符算 2）
func runeWidth(r rune) int {
	if r > 0x2E7F {
		return 2
	}
	return 1
}

// 把一行按词折行，词内不打断；中文等连写文本没有空格时按字符宽度硬切
func wrapLine(line string, width int) []string {
	if width <= 0 {
		return []string{line}
	}
	var out []string
	var cur strings.Builder
	curW := 0
	flush := func() {
		if cur.Len() > 0 {
			out = append(out, cur.String())
			cur.Reset()
			curW = 0
		}
	}
	words := strings.Fields(line)
	if len(words) == 0 {
		return []string{""}
	}
	for i, w := range words {
		wW := 0
		for _, r := range w {
			wW += runeWidth(r)
		}
		// 单个词比整行还宽，只能硬切
		if wW > width {
			flush()
			var piece strings.Builder
			pieceW := 0
			for _, r := range w {
				rw := runeWidth(r)
				if pieceW+rw > width && piece.Len() > 0 {
					out = append(out, piece.String())
					piece.Reset()
					pieceW = 0
				}
				piece.WriteRune(r)
				pieceW += rw
			}
			cur.WriteString(piece.String())
			curW += pieceW
			continue
		}
		// 词间空格也要占宽度
		sep := 0
		if i > 0 {
			sep = 1
		}
		if curW+sep+wW > width && cur.Len() > 0 {
			flush()
		}
		if sep > 0 && cur.Len() > 0 {
			cur.WriteByte(' ')
			curW += 1
		}
		cur.WriteString(w)
		curW += wW
	}
	flush()
	return out
}

// wrapText 多行文本整体折行
func wrapText(text string, width int, indent, prefix string) string {
	var b strings.Builder
	sc := bufio.NewScanner(strings.NewReader(text))
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	n := 0
	for sc.Scan() {
		n++
		head := prefix
		if head == "" && indent != "" {
			head = indent
		} else if head != "" && indent != "" {
			head = indent + head
		}
		w := width - displayWidth(head)
		if w < 1 {
			w = 1
		}
		for i, l := range wrapLine(sc.Text(), w) {
			if i == 0 {
				b.WriteString(head)
			} else {
				b.WriteString(indent)
			}
			b.WriteString(l)
			b.WriteByte('\n')
		}
	}
	return b.String()
}

// displayWidth 算一段纯 ASCII 前缀的显示宽度（前缀一般不含宽字符）
func displayWidth(s string) int {
	return len(s)
}

func main() {
	width := flag.Int("w", 80, "折行宽度，中文按两个宽度计")
	indent := flag.String("i", "", "每行缩进，比如两个空格或制表符")
	prefix := flag.String("t", "", "行前缀，会放在缩进之后、正文之前")
	flag.Parse()

	text := readStdin()
	if len(flag.Args()) > 0 {
		// 也支持直接把文件当参数读进来折行
		var b strings.Builder
		for _, p := range flag.Args() {
			d, err := os.ReadFile(p)
			if err != nil {
				fmt.Fprintf(os.Stderr, "读 %s 失败: %v\n", p, err)
				continue
			}
			b.Write(d)
			b.WriteByte('\n')
		}
		text = b.String()
	}
	fmt.Print(wrapText(text, *width, *indent, *prefix))
}

func readStdin() string {
	b, err := io.ReadAll(os.Stdin)
	if err != nil {
		return ""
	}
	return string(b)
}
