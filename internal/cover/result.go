package cover

import (
	"sort"

	"task239-lemmareview/internal/model"
)

// StepPremiseLemmas 由前提边聚合出「步骤 → 其依赖的引理 ID 列表」映射，
// 供豁免裁决引擎判断某缺失步骤的前提是否可被整体豁免。
func StepPremiseLemmas(edges []*model.PremiseEdge) map[int64][]int64 {
	out := make(map[int64][]int64)
	for _, e := range edges {
		if e.FromKind != "lemma" {
			continue
		}
		out[e.ToStepID] = append(out[e.ToStepID], e.FromID)
	}
	for k := range out {
		sort.Slice(out[k], func(i, j int) bool { return out[k][i] < out[k][j] })
	}
	return out
}
