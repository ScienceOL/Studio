package utils

import "errors"

type Node[T comparable, D any] struct {
	Name   T
	Parent T
	Data   D
}

func BuildDAGHierarchy[T comparable, D any](nodes []*Node[T, D]) ([][]*Node[T, D], error) {
	if len(nodes) == 0 {
		return [][]*Node[T, D]{}, nil
	}

	// 获取零值用于判断根节点
	var zeroValue T

	// 构建邻接表和入度表
	graph := make(map[T][]*Node[T, D])

	nodeMap := SliceToMap(nodes, func(item *Node[T, D]) (T, *Node[T, D]) {
		return item.Name, item
	})

	if len(nodes) != len(nodeMap) {
		return [][]*Node[T, D]{}, errors.New("节点名重复")
	}

	inDegree := SliceToMap(nodes, func(item *Node[T, D]) (T, int) {
		return item.Name, 0
	})

	// 构建图和计算入度
	for _, node := range nodes {
		// 如果父节点不存在，则不添加
		if _, ok := nodeMap[node.Parent]; !ok {
			continue
		}

		if node.Parent != zeroValue {
			graph[node.Parent] = append(graph[node.Parent], node)
			inDegree[node.Name]++
		}
	}

	// 使用Kahn算法进行拓扑排序
	var levels [][]*Node[T, D]
	processed := make(map[T]bool)

	queues := MapToSlice(inDegree, func(k T, v int) (*Node[T, D], bool) {
		if v == 0 {
			node, _ := nodeMap[k]
			return node, true
		}
		return nil, false
	})

	for len(queues) > 0 {
		var currentLevel []*Node[T, D]
		var nextQueue []*Node[T, D]

		// 处理当前层级的所有节点
		for _, node := range queues {
			processed[node.Name] = true

			// 只有在原始节点列表中的节点才添加到层级中
			if _, exists := nodeMap[node.Name]; exists {
				currentLevel = append(currentLevel, node)
			}

			// 处理子节点
			for _, childNode := range graph[node.Name] {
				inDegree[childNode.Name]--
				if inDegree[childNode.Name] == 0 {
					nextQueue = append(nextQueue, childNode)
				}
			}
		}

		if len(currentLevel) > 0 {
			levels = append(levels, currentLevel)
		}

		queues = nextQueue
	}

	// 检查是否所有原始节点都被处理（如果没有，说明有环）
	for _, node := range nodes {
		if !processed[node.Name] {
			return nil, errors.New("图中存在环，这不是一个有效的DAG")
		}
	}

	return levels, nil
}
