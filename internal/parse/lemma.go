package parse

import (
	"fmt"
	"strings"
)

// LemmaSpec 解析出的引理声明（尚未落库）。
type LemmaSpec struct {
	Name      string
	Statement string
}

// ParseLemmaBlock 解析引理声明块。每行「名称: 陈述」，空行分隔不同引理。
func ParseLemmaBlock(text string) ([]LemmaSpec, error) {
	lines := strings.Split(text, "\n")
	var out []LemmaSpec
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		name := line
		stmt := ""
		if idx := strings.Index(line, ":"); idx >= 0 {
			name = strings.TrimSpace(line[:idx])
			stmt = strings.TrimSpace(line[idx+1:])
		}
		if name == "" {
			return nil, fmt.Errorf("empty lemma name in line: %q", line)
		}
		out = append(out, LemmaSpec{Name: name, Statement: stmt})
	}
	return out, nil
}
