// Package adjudicate 处理假设豁免裁决：研究者可针对某步骤对某前提声明显式豁免，
// 豁免后的缺失前提不再阻断该步骤覆盖，但豁免须记录在案并可追溯。
package adjudicate

import (
	"fmt"
	"sort"

	"task239-lemmareview/internal/model"
)

// Engine 豁免裁决引擎。
type Engine struct {
	exemptions []*model.Exemption
}

// New 构造引擎。
func New(exemptions []*model.Exemption) *Engine {
	return &Engine{exemptions: exemptions}
}

// IsExempt 判断某步骤对某引理是否已被豁免。
func (e *Engine) IsExempt(stepID, lemmaID int64) bool {
	for _, ex := range e.exemptions {
		if ex.StepID == stepID && ex.LemmaID == lemmaID {
			return true
		}
	}
	return false
}

// ExemptReasons 返回某步骤的全部豁免理由。
func (e *Engine) ExemptReasons(stepID int64) []string {
	var out []string
	for _, ex := range e.exemptions {
		if ex.StepID == stepID {
			out = append(out, fmt.Sprintf("lemma=%d: %s", ex.LemmaID, ex.Reason))
		}
	}
	sort.Strings(out)
	return out
}

// ApplyExemptions 在覆盖分析后，将豁免的步骤从缺失列表移出（标注为豁免覆盖）。
// 返回更新后的缺失列表与被豁免的步骤列表。
func (e *Engine) ApplyExemptions(missing []int64, stepToLemmas map[int64][]int64) (stillMissing []int64, exempted []int64) {
	for _, sid := range missing {
		lemmas := stepToLemmas[sid]
		allExempt := true
		for _, lid := range lemmas {
			if !e.IsExempt(sid, lid) {
				allExempt = false
				break
			}
		}
		if allExempt && len(lemmas) > 0 {
			exempted = append(exempted, sid)
		} else {
			stillMissing = append(stillMissing, sid)
		}
	}
	return stillMissing, exempted
}
