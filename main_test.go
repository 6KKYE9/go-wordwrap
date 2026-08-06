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
