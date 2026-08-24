// Package graph 管理证明依赖图：以推导步骤为节点、前提依赖边为有向边。
// 本文件负责图的构建；循环检测见 cycle.go。
package graph

import (
	"task239-lemmareview/internal/model"
)

// DepGraph 依赖图：节点为步骤 ID，边 from→to 表示 to 依赖 from（from 是前提）。
type DepGraph struct {
	nodes map[int64]bool
	// adj[from] = set of to（from 是前提，to 依赖它）
	adj map[int64]map[int64]bool
	// revAdj[to] = set of from（to 依赖的前提来源）
	revAdj map[int64]map[int64]bool
}

// Build 由步骤列表与前提边构造依赖图。
func Build(steps []*model.Step, edges []*model.PremiseEdge) *DepGraph {
	g := &DepGraph{
		nodes:  make(map[int64]bool),
		adj:    make(map[int64]map[int64]bool),
		revAdj: make(map[int64]map[int64]bool),
	}
	for _, s := range steps {
		g.nodes[s.ID] = true
	}
	for _, e := range edges {
		// from 为被依赖方（前提来源），to 为依赖方（步骤本身）
		from := e.FromID
		if e.FromKind == "step" {
			g.nodes[from] = true
		}
		to := e.ToStepID
		if g.adj[from] == nil {
			g.adj[from] = make(map[int64]bool)
		}
		g.adj[from][to] = true
		if g.revAdj[to] == nil {
			g.revAdj[to] = make(map[int64]bool)
		}
		g.revAdj[to][from] = true
	}
	return g
}

// HasNode 判断节点是否存在。
func (g *DepGraph) HasNode(id int64) bool { return g.nodes[id] }

// Adj 返回某节点的出边目标集合（依赖它的步骤）。
func (g *DepGraph) Adj(from int64) map[int64]bool {
	return g.adj[from]
}

// RevAdj 返回某节点依赖的前提来源集合。
func (g *DepGraph) RevAdj(to int64) map[int64]bool {
	return g.revAdj[to]
}
