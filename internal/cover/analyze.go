// Package cover 实现前提可达性传播与覆盖分析：
// 从「可用引理」和「首次步骤（无前提依赖）」出发，沿依赖图正向传播已成立的结论，
// 标记每个推导步骤的前提是否全部满足（覆盖），否则标记为缺失或循环相关。
package cover

import (
	"sort"

	"task239-lemmareview/internal/model"
)

// Analyzer 覆盖分析器。
type Analyzer struct {
	steps  []*model.Step
	edges  []*model.PremiseEdge
	lemmas map[int64]*model.Lemma // 可用引理集合（仅 available 状态计入前提源）
}

// New 构造分析器。availableLemmas 为状态为 available 的引理。
func New(steps []*model.Step, edges []*model.PremiseEdge, availableLemmas []*model.Lemma) *Analyzer {
	m := make(map[int64]*model.Lemma)
	for _, l := range availableLemmas {
		if l.Status == model.LemmaAvailable {
			m[l.ID] = l
		}
	}
	return &Analyzer{steps: steps, edges: edges, lemmas: m}
}

// Analyze 执行覆盖分析。
// 算法：
//  1. 初始已成立前提集合 = 所有可用引理 ID（from_kind=lemma）。
//  2. 将无前提依赖的步骤视为「已覆盖」（其结论可作为后续前提）。
//  3. 迭代：反复扫描步骤，若某步骤的全部前提（引理或前置步骤结论）均已成立，则标记覆盖并加入已成立集合。
//  4. 固定点后仍未覆盖且非循环的步骤 → 缺失；被循环检测标记的步骤 → 循环相关。
func (a *Analyzer) Analyze(cyclic map[int64]bool) *model.CoverageResult {
	established := make(map[int64]bool) // 已成立的前提源 ID（引理或步骤）
	for lid := range a.lemmas {
		established[lid] = true
	}

	stepByID := make(map[int64]*model.Step, len(a.steps))
	premisesOf := make(map[int64][]model.PremiseEdge) // stepID -> 前提边
	for _, s := range a.steps {
		stepByID[s.ID] = s
	}
	for _, e := range a.edges {
		premisesOf[e.ToStepID] = append(premisesOf[e.ToStepID], *e)
	}

	// 初始：无前提依赖的步骤直接覆盖
	covered := make(map[int64]bool)
	for _, s := range a.steps {
		if len(premisesOf[s.ID]) == 0 {
			covered[s.ID] = true
			established[s.ID] = true // 其结论成为可用前提
		}
	}

	// 迭代至固定点
	changed := true
	for changed {
		changed = false
		for _, s := range a.steps {
			if covered[s.ID] || cyclic[s.ID] {
				continue
			}
			prems := premisesOf[s.ID]
			if len(prems) == 0 {
				continue
			}
			allMet := true
			for _, p := range prems {
				var ok bool
				if p.FromKind == "lemma" {
					_, ok = a.lemmas[p.FromID]
				} else {
					ok = established[p.FromID]
				}
				// 仅必修前提未满足才阻断覆盖；可选前提未满足不影响该步骤成立。
				if !ok && p.Required {
					allMet = false
					break
				}
			}
			if allMet {
				covered[s.ID] = true
				established[s.ID] = true
				changed = true
			}
		}
	}

	res := &model.CoverageResult{}
	res.AvailablePremises = sortedKeys(a.lemmas)

	for _, s := range a.steps {
		switch {
		case cyclic[s.ID]:
			res.CyclicSteps = append(res.CyclicSteps, s.ID)
		case covered[s.ID]:
			res.CoveredSteps = append(res.CoveredSteps, s.ID)
		default:
			res.MissingSteps = append(res.MissingSteps, s.ID)
		}
	}
	sort.Slice(res.CoveredSteps, func(i, j int) bool { return res.CoveredSteps[i] < res.CoveredSteps[j] })
	sort.Slice(res.MissingSteps, func(i, j int) bool { return res.MissingSteps[i] < res.MissingSteps[j] })
	sort.Slice(res.CyclicSteps, func(i, j int) bool { return res.CyclicSteps[i] < res.CyclicSteps[j] })
	return res
}

func sortedKeys(m map[int64]*model.Lemma) []int64 {
	out := make([]int64, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}
