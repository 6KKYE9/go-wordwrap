package main

import "testing"

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
	out := wrapText("ab cd ef", 4)
	if out != "ab\ncd\nef\n" {
		t.Fatalf("多行整体折行不符: %q", out)
	}
}
