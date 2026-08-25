// Package model 定义数学证明引理依赖复核台的核心领域实体、状态枚举与错误类型。
package model

import (
	"errors"
	"fmt"
	"time"
)

// 状态枚举：证明草稿
type DraftStatus string

const (
	DraftEditing     DraftStatus = "editing"     // 编辑中
	DraftReviewing   DraftStatus = "reviewing"   // 待复核
	DraftGap         DraftStatus = "gap"         // 存在缺口
	DraftPublishable DraftStatus = "publishable" // 可发布
	DraftFrozen      DraftStatus = "frozen"      // 冻结
)

// 状态枚举：推导步骤
type StepStatus string

const (
	StepParsed  StepStatus = "parsed"  // 待解析（已解析）
	StepCovered StepStatus = "covered" // 已覆盖
	StepMissing StepStatus = "missing" // 前提缺失
	StepCyclic  StepStatus = "cyclic"  // 循环相关
)

// 状态枚举：引理
type LemmaStatus string

const (
	LemmaCandidate LemmaStatus = "candidate" // 候选
	LemmaAvailable LemmaStatus = "available" // 可用
	LemmaReplaced  LemmaStatus = "replaced"  // 被替代
	LemmaInvalid   LemmaStatus = "invalid"   // 失效
)

// 状态枚举：证明版本
type VersionStatus string

const (
	VersionDraft      VersionStatus = "draft"      // 草稿
	VersionShared     VersionStatus = "shared"     // 共享
	VersionFrozen     VersionStatus = "frozen"     // 冻结
	VersionSuperseded VersionStatus = "superseded" // 替代
)

// Draft 证明草稿
type Draft struct {
	ID          int64       `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Status      DraftStatus `json:"status"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// Step 推导步骤
type Step struct {
	ID         int64      `json:"id"`
	DraftID    int64      `json:"draft_id"`
	Seq        int        `json:"seq"`        // 推导顺序
	Label      string     `json:"label"`      // 步骤标号，如 "2.3"
	Statement  string     `json:"statement"`  // 步骤陈述
	Conclusion string     `json:"conclusion"` // 该步导出的结论（命题）
	Status     StepStatus `json:"status"`
	CreatedAt  time.Time  `json:"created_at"`
}

// Lemma 引理 / 前提
type Lemma struct {
	ID        int64       `json:"id"`
	DraftID   int64       `json:"draft_id"`
	Name      string      `json:"name"`      // 引理名
	Statement string      `json:"statement"` // 引理陈述
	Status    LemmaStatus `json:"status"`
	CreatedAt time.Time   `json:"created_at"`
}

// PremiseEdge 前提依赖边：某步骤的结论依赖某引理（或依赖另一步骤的结论）
type PremiseEdge struct {
	ID        int64     `json:"id"`
	DraftID   int64     `json:"draft_id"`
	FromKind  string    `json:"from_kind"`  // "step" | "lemma"
	FromID    int64     `json:"from_id"`    // 被依赖方（前提来源）
	ToStepID  int64     `json:"to_step_id"` // 依赖方（推导步骤）
	Required  bool      `json:"required"`   // 是否必修前提
	CreatedAt time.Time `json:"created_at"`
}

// Exemption 假设豁免：研究者声明某步骤的某前提被显式豁免
type Exemption struct {
	ID        int64     `json:"id"`
	DraftID   int64     `json:"draft_id"`
	StepID    int64     `json:"step_id"`
	LemmaID   int64     `json:"lemma_id"` // 被豁免的引理（0 表示基于步骤）
	Reason    string    `json:"reason"`
	CreatedAt time.Time `json:"created_at"`
}

// ProofVersion 证明版本快照
type ProofVersion struct {
	ID        int64         `json:"id"`
	DraftID   int64         `json:"draft_id"`
	Name      string        `json:"name"`
	Status    VersionStatus `json:"status"`
	LemmaSet  string        `json:"lemma_set"` // 冻结时绑定的引理集合指纹
	Snapshot  string        `json:"snapshot"`  // 步骤/前提图快照
	CreatedAt time.Time     `json:"created_at"`
}

// CoverageResult 前提覆盖分析结果（内存计算，不持久化原始计算）
type CoverageResult struct {
	DraftID           int64     `json:"draft_id"`
	CoveredSteps      []int64   `json:"covered_steps"`
	ExemptedSteps     []int64   `json:"exempted_steps"`
	MissingSteps      []int64   `json:"missing_steps"`
	CyclicSteps       []int64   `json:"cyclic_steps"`
	Cycles            [][]int64 `json:"cycles"` // 检测到的循环（步骤 ID 序列）
	AvailablePremises []int64   `json:"available_premises"`
}

// 错误定义
var (
	ErrDraftNotFound   = errors.New("draft not found")
	ErrStepNotFound    = errors.New("step not found")
	ErrLemmaNotFound   = errors.New("lemma not found")
	ErrSelfPremise     = errors.New("self-referencing premise: a step cannot depend on its own conclusion")
	ErrUnknownLemma    = errors.New("unknown lemma referenced by premise edge")
	ErrOrderConflict   = errors.New("derivation order conflict: dependency precedes or equals dependent step")
	ErrFrozenWrite     = errors.New("cannot modify a frozen draft or version")
	ErrVersionConflict = errors.New("concurrent edit conflict: step version mismatch")
	ErrDuplicateStep   = errors.New("duplicate step label in same draft")
	ErrInvalidStatus   = errors.New("invalid status transition")
)

// ValidateStepOrder 校验依赖边的推导顺序合法性。
// dependent 为依赖方步骤的顺序号，required 为被依赖方步骤的顺序号（引理依赖传 -1 表示不约束顺序）。
// 仅当被依赖方序号严格大于依赖方（即前提出现在结论之后，逻辑上不可能前向推导）时报错；
// 允许 dependent > required（正常前向）或 dependent == required（可能构成同序逻辑环，由图算法检测）。
func ValidateStepOrder(dependentSeq, requiredSeq int) error {
	if requiredSeq > dependentSeq {
		return fmt.Errorf("%w (required seq=%d must not exceed dependent seq=%d)", ErrOrderConflict, requiredSeq, dependentSeq)
	}
	return nil
}

// ValidDraftTransition 判断草稿状态机是否允许从 current 进入 next。
func ValidDraftTransition(current, next DraftStatus) bool {
	if current == next {
		return true
	}
	if current == DraftFrozen {
		return true
	}
	switch current {
	case DraftEditing:
		return next == DraftReviewing
	case DraftReviewing:
		return next == DraftGap || next == DraftPublishable
	case DraftGap:
		return next == DraftReviewing || next == DraftPublishable
	case DraftPublishable:
		return next == DraftReviewing || next == DraftGap || next == DraftFrozen
	default:
		return false
	}
}
