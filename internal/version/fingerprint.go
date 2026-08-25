// Package version 负责证明版本快照的指纹生成与冻结逻辑：
// 冻结时绑定当前草稿的引理集合指纹与步骤/前提图快照，保证版本不可变且可复核。
// 本文件计算引理集合指纹。
package version

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"task239-lemmareview/internal/model"
)

// LemmaSetFingerprint 计算引理集合指纹：按 ID 排序后拼接状态，取 SHA-256。
func LemmaSetFingerprint(lemmas []*model.Lemma) string {
	ids := make([]int64, 0, len(lemmas))
	statusByID := make(map[int64]string, len(lemmas))
	for _, l := range lemmas {
		ids = append(ids, l.ID)
		statusByID[l.ID] = string(l.Status)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	var b strings.Builder
	for _, id := range ids {
		b.WriteString(fmt.Sprintf("%d:%s;", id, statusByID[id]))
	}
	sum := sha256.Sum256([]byte(b.String()))
	return hex.EncodeToString(sum[:])
}
