// Package parse 提供推导步骤文本解析与引理声明解析能力。
// 本文件处理推导步骤：从研究者提交的步骤文本中抽出序号、标号、陈述与结论。
package parse

import (
	"fmt"
	"regexp"
	"strings"

	"task239-lemmareview/internal/model"
)

// StepSpec 解析出的步骤规格（尚未落库）。
type StepSpec struct {
	Seq        int
	Label      string
	Statement  string
	Conclusion string
}

var labelRe = regexp.MustCompile(`^\s*(\d+(?:\.\d+)*)\b`)

// ParseSteps 解析多行步骤文本。每行应为「标号 陈述 => 结论」或以「标号」开头。
// 结论部分可选，若含 "=>" 则右侧为结论。空行忽略。
func ParseSteps(text string) ([]StepSpec, error) {
	lines := strings.Split(text, "\n")
	var out []StepSpec
	seq := 0
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		m := labelRe.FindStringSubmatch(line)
		if m == nil {
			return nil, fmt.Errorf("cannot find step label in line: %q", line)
		}
		label := m[1]
		rest := strings.TrimSpace(line[len(m[0]):])
		statement := rest
		conclusion := ""
		if idx := strings.Index(rest, "=>"); idx >= 0 {
			statement = strings.TrimSpace(rest[:idx])
			conclusion = strings.TrimSpace(rest[idx+2:])
		}
		seq++
		out = append(out, StepSpec{
			Seq:        seq,
			Label:      label,
			Statement:  statement,
			Conclusion: conclusion,
		})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no valid step parsed")
	}
	return out, nil
}

// ValidateLabelsUnique 校验同一草稿内步骤标号唯一。
func ValidateLabelsUnique(specs []StepSpec) error {
	seen := make(map[string]bool, len(specs))
	for _, s := range specs {
		if seen[s.Label] {
			return fmt.Errorf("%w: label %q duplicated", model.ErrDuplicateStep, s.Label)
		}
		seen[s.Label] = true
	}
	return nil
}
