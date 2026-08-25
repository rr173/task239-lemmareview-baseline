package version

import (
	"fmt"
	"sort"
	"strings"

	"task239-lemmareview/internal/model"
)

// GraphSnapshot 生成步骤/前提图的确定性快照字符串。
func GraphSnapshot(steps []*model.Step, edges []*model.PremiseEdge) string {
	type stepLine struct {
		id    int64
		seq   int
		label string
	}
	sl := make([]stepLine, 0, len(steps))
	for _, s := range steps {
		sl = append(sl, stepLine{id: s.ID, seq: s.Seq, label: s.Label})
	}
	sort.Slice(sl, func(i, j int) bool { return sl[i].id < sl[j].id })

	type edgeLine struct {
		fromKind string
		fromID   int64
		toStepID int64
	}
	el := make([]edgeLine, 0, len(edges))
	for _, e := range edges {
		el = append(el, edgeLine{fromKind: e.FromKind, fromID: e.FromID, toStepID: e.ToStepID})
	}
	sort.Slice(el, func(i, j int) bool {
		if el[i].fromKind != el[j].fromKind {
			return el[i].fromKind < el[j].fromKind
		}
		if el[i].fromID != el[j].fromID {
			return el[i].fromID < el[j].fromID
		}
		return el[i].toStepID < el[j].toStepID
	})

	var b strings.Builder
	for _, s := range sl {
		b.WriteString(fmt.Sprintf("S[%d|%d|%s]\n", s.id, s.seq, s.label))
	}
	for _, e := range el {
		b.WriteString(fmt.Sprintf("E[%s|%d|%d]\n", e.fromKind, e.fromID, e.toStepID))
	}
	return b.String()
}

// Freeze 生成冻结版本所需字段：状态置为 frozen，绑定引理指纹与图快照。
func Freeze(name string, lemmas []*model.Lemma, steps []*model.Step, edges []*model.PremiseEdge) *model.ProofVersion {
	return &model.ProofVersion{
		Name:      name,
		Status:    model.VersionFrozen,
		LemmaSet:  LemmaSetFingerprint(lemmas),
		Snapshot:  GraphSnapshot(steps, edges),
	}
}
