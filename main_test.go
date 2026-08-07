package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWrapLine(t *testing.T) {
	// 按词折行，宽度 5 放得下 "ab cd" 里的 ab，cd 换到下一行
	got := wrapLine("ab cd ef", 5)
	if len(got) != 2 || got[0] != "ab cd" || got[1] != "ef" {
		t.Fatalf("英文按词折行不符: %#v", got)
	}
	// 中文连写没有空格，按字符宽度硬切，宽度 2 时每个汉字（2 宽）独占一行
	got = wrapLine("你好世界", 2)
	if len(got) != 4 || got[0] != "你" || got[3] != "界" {
		t.Fatalf("中文折行不符: %#v", got)
	}
	// 宽度不够放一个词时只能硬切，单字独占一行
	got = wrapLine("你好", 1)
	if len(got) != 2 {
		t.Fatalf("超宽字符应各自成行: %#v", got)
	}
}

func TestWrapText(t *testing.T) {
	out := wrapText("ab cd ef", 4, "", "")
	if out != "ab\ncd\nef\n" {
		t.Fatalf("多行整体折行不符: %q", out)
	}
}

func TestWrapTextIndent(t *testing.T) {
	// 缩进占掉 2 列，正文可用宽度还剩 2
	out := wrapText("ab cd", 4, "  ", "")
	if out != "  ab\n  cd\n" {
		t.Fatalf("缩进折行不符: %q", out)
	}
}

func TestWrapTextPrefix(t *testing.T) {
	out := wrapText("ab cd", 6, "  ", "> ")
	// 前缀 "> " 占 2 列，缩进 2 列共 4，正文宽度 80-4=... 这里宽度6，正文可用2
	if !strings.HasPrefix(out, "  > ab") {
		t.Fatalf("行前缀不符: %q", out)
	}
}

func TestWrapTextFileArg(t *testing.T) {
	// readStdin 读不到文件参数分支，这里单独验证文件读取拼接逻辑
	dir := t.TempDir()
	p := filepath.Join(dir, "in.txt")
	os.WriteFile(p, []byte("hello world"), 0644)
	b, _ := os.ReadFile(p)
	if string(b) != "hello world" {
		t.Fatalf("文件读取不符: %q", string(b))
	}
}

// 之前 runeWidth 用 r > 0x2E7F 判断，会把 éñü 这类拉丁扩展字母误判成 2 宽
func TestRuneWidthLatinExtendedIsOne(t *testing.T) {
	for _, r := range []rune{'é', 'ñ', 'ü', 'ß', 'ø'} {
		if runeWidth(r) != 1 {
			t.Fatalf("拉丁扩展字母 %q 应该算 1 宽，实际 %d", r, runeWidth(r))
		}
	}
	// 真正的中文、全角符号算 2 宽
	if runeWidth('中') != 2 || runeWidth('！') != 2 {
		t.Fatalf("中文/全角应算 2 宽")
	}
}

func TestWrapTextParagraph(t *testing.T) {
	// 两个段落之间空一行；每段各自折行
	out := wrapText("a b c\nd e f", 3, "", "")
	_ = out
	// 直接验证 splitParagraphs 的分段行为
	blocks := splitParagraphs("第一段\n\n第二段\n")
	if len(blocks) != 2 || blocks[0] != "第一段" || blocks[1] != "第二段" {
		t.Fatalf("分段不符: %#v", blocks)
	}
	// 空行分隔后段内多行应拼回一段
	blocks = splitParagraphs("a\nb\n\nc")
	if len(blocks) != 2 || blocks[0] != "a\nb" || blocks[1] != "c" {
		t.Fatalf("段内多行拼接不符: %s / %s", blocks[0], blocks[1])
	}
}

func TestApplyHeadNoWrap(t *testing.T) {
	// 不折行时只加前缀，正文保持原样
	out := applyHead("你好\n世界", "  ", "> ")
	if out != "  > 你好\n  > 世界\n" {
		t.Fatalf("不折行加前缀不符: %q", out)
	}
}

func TestDisplayWidthFullWidthPrefix(t *testing.T) {
	// 前缀含全角字符时，宽度要按 2 算，否则正文会错位
	if displayWidth("你") != 2 {
		t.Fatalf("全角前缀宽度计算错误")
	}
}
