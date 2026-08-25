package graph

import "fmt"

// DetectCycles 使用 DFS 三色标记法检测有向环，返回所有找到的环路（步骤 ID 序列）。
func (g *DepGraph) DetectCycles() ([][]int64, error) {
	const (
		white = 0
		gray  = 1
		black = 2
	)
	color := make(map[int64]int, len(g.nodes))
	for n := range g.nodes {
		color[n] = white
	}
	var cycles [][]int64
	var stack []int64

	var dfs func(int64) error
	dfs = func(u int64) error {
		color[u] = gray
		stack = append(stack, u)
		for v := range g.adj[u] {
			if !g.nodes[v] {
				continue
			}
			switch color[v] {
			case white:
				if err := dfs(v); err != nil {
					return err
				}
			case gray:
				// 找到回边，提取环路
				start := -1
				for i, n := range stack {
					if n == v {
						start = i
						break
					}
				}
				if start < 0 {
					return fmt.Errorf("internal: cycle start not found")
				}
				cycle := append([]int64{}, stack[start:]...)
				cycle = append(cycle, v) // 闭合
				cycles = append(cycles, cycle)
			case black:
				// 已完成的子树，忽略
			}
		}
		color[u] = black
		stack = stack[:len(stack)-1]
		return nil
	}

	for n := range g.nodes {
		if color[n] == white {
			if err := dfs(n); err != nil {
				return nil, err
			}
		}
	}
	return cycles, nil
}

// DirectCyclicSteps 返回直接自检（步骤依赖自身结论）的步骤集合。
func (g *DepGraph) DirectCyclicSteps() map[int64]bool {
	out := make(map[int64]bool)
	for from, toSet := range g.adj {
		if toSet[from] {
			out[from] = true
		}
	}
	return out
}
