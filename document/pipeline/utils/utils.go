package utils

import (
	"github.com/wsnacj/agentx-go/document/pipeline/section"
	"strings"
)

func JoinAndClip(pages []string, max int) string {
	text := strings.Join(pages, "\n")
	rs := []rune(text)
	if max > 0 && len(rs) > max {
		return string(rs[:max])
	}
	return text
}

func Flatten(nodes []*section.Node) []*section.Node {
	var out []*section.Node
	var dfs func(*section.Node)
	dfs = func(n *section.Node) {
		out = append(out, n)
		for _, c := range n.Children {
			dfs(c)
		}
	}
	for _, n := range nodes {
		dfs(n)
	}
	return out
}

func UniqueKeepOrder(in []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(in))
	for _, v := range in {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}

func TakePages(pages []string, idxs []int) []string {
	var out []string
	for _, pi := range idxs {
		if pi >= 1 && pi <= len(pages) {
			out = append(out, pages[pi-1])
		}
	}
	return out
}

func Ternary(cond bool, a, b string) string {
	if cond {
		return a
	}
	return b
}
