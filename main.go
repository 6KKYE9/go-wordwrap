package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

// 一个字符占几列宽。中文、日文、韩文、全角符号这些算 2 宽；
// 之前用 r > 0x2E7F 一刀切，会把 éñü 这类拉丁扩展字母也误判成 2 宽
func runeWidth(r rune) int {
	if r == 0 {
		return 0
	}
	// 全角及半角形、全角符号
	if r >= 0xFF00 && r <= 0xFF60 {
		return 2
	}
	if r >= 0xFFE0 && r <= 0xFFE6 {
		return 2
	}
	// CJK 统一表意文字、扩展区
	if r >= 0x1100 && r <= 0x115F {
		return 2
	}
	if r >= 0x2E80 && r <= 0x303E {
		return 2
	}
	if r >= 0x3041 && r <= 0x33FF {
		return 2
	}
	if r >= 0x3400 && r <= 0x4DBF {
		return 2
	}
	if r >= 0x4E00 && r <= 0x9FFF {
		return 2
	}
	if r >= 0xA000 && r <= 0xA4CF {
		return 2
	}
	if r >= 0xAC00 && r <= 0xD7A3 {
		return 2
	}
	if r >= 0xF900 && r <= 0xFAFF {
		return 2
	}
	if r >= 0xFE30 && r <= 0xFE4F {
		return 2
	}
	if r >= 0x20000 && r <= 0x3FFFD {
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
	for sc.Scan() {
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

// displayWidth 算一段前缀的真实显示宽度（前缀里也可能含全角字符）
func displayWidth(s string) int {
	n := 0
	for _, r := range s {
		n += runeWidth(r)
	}
	return n
}

func main() {
	width := flag.Int("w", 80, "折行宽度，中文按两个宽度计；0 表示不折行")
	indent := flag.String("i", "", "每行缩进，比如两个空格或制表符")
	prefix := flag.String("t", "", "行前缀，会放在缩进之后、正文之前")
	para := flag.Bool("s", false, "按段落折行：两段之间空一行（按空行分段）")
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
	if *width <= 0 {
		// 不折行就原样输出，但缩进和前缀还是给上
		fmt.Print(applyHead(text, *indent, *prefix))
		return
	}
	if *para {
		// 按空行把文本切成段落，每段各自折行、段间留空行
		var b strings.Builder
		for _, blk := range splitParagraphs(text) {
			b.WriteString(wrapText(blk, *width, *indent, *prefix))
			b.WriteByte('\n')
		}
		fmt.Print(b.String())
		return
	}
	fmt.Print(wrapText(text, *width, *indent, *prefix))
}

// applyHead 给每一行加缩进/前缀但不折行
func applyHead(text, indent, prefix string) string {
	var b strings.Builder
	sc := bufio.NewScanner(strings.NewReader(text))
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		head := prefix
		if head == "" && indent != "" {
			head = indent
		} else if head != "" && indent != "" {
			head = indent + head
		}
		b.WriteString(head)
		b.WriteString(sc.Text())
		b.WriteByte('\n')
	}
	return b.String()
}

// splitParagraphs 按连续空行把文本切成段落，段末的空行不单独成段
func splitParagraphs(text string) []string {
	raw := strings.Split(text, "\n")
	var blocks []string
	for _, blk := range raw {
		// 整行都是空白的行当成段落分隔
		if strings.TrimSpace(blk) == "" {
			// 已经在段落分隔（上一个块是占位）就不重复加
			if len(blocks) == 0 || blocks[len(blocks)-1] == "" {
				continue
			}
			blocks = append(blocks, "")
			continue
		}
		if len(blocks) > 0 && blocks[len(blocks)-1] == "" {
			// 接在空占位后面，开始一个新段落
			blocks[len(blocks)-1] = blk
		} else if len(blocks) == 0 {
			blocks = append(blocks, blk)
		} else {
			// 和上一行同属一段，拼回去
			blocks[len(blocks)-1] = blocks[len(blocks)-1] + "\n" + blk
		}
	}
	if len(blocks) == 0 {
		return []string{""}
	}
	// 去掉末尾因为结尾换行多出来的空段落
	for len(blocks) > 0 && blocks[len(blocks)-1] == "" {
		blocks = blocks[:len(blocks)-1]
	}
	return blocks
}

func readStdin() string {
	b, err := io.ReadAll(os.Stdin)
	if err != nil {
		return ""
	}
	return string(b)
}
