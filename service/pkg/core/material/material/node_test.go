package material

import (
	"errors"
	"fmt"
)

// Node 输入节点结构
type Node struct {
	Name       string
	ParentName string
}

// LevelNode 层级节点结构
type LevelNode struct {
	Name       string
	ParentName string
}

// Level 层级结构
type Level struct {
	Level int
	Nodes []LevelNode
}

// detectCycleAndGetLevels 检测DAG是否有环并返回层级结构
func detectCycleAndGetLevels(nodes []Node) ([]Level, error) {
	if len(nodes) == 0 {
		return []Level{}, nil
	}

	// 构建邻接表和入度表
	graph := make(map[string][]string)
	inDegree := make(map[string]int)
	nodeMap := make(map[string]Node)
	allNodes := make(map[string]bool)

	// 初始化所有节点
	for _, node := range nodes {
		nodeMap[node.Name] = node
		allNodes[node.Name] = true
		if node.ParentName != "" {
			allNodes[node.ParentName] = true
		}
		inDegree[node.Name] = 0
	}

	// 为所有节点（包括父节点）初始化入度
	for nodeName := range allNodes {
		if _, exists := inDegree[nodeName]; !exists {
			inDegree[nodeName] = 0
		}
	}

	// 构建图和计算入度
	for _, node := range nodes {
		if node.ParentName != "" {
			graph[node.ParentName] = append(graph[node.ParentName], node.Name)
			inDegree[node.Name]++
		}
	}

	// 使用Kahn算法进行拓扑排序
	queue := []string{}
	levels := []Level{}
	processed := make(map[string]bool)

	// 找到所有入度为0的节点（根节点）
	for nodeName, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, nodeName)
		}
	}

	levelNum := 0
	for len(queue) > 0 {
		currentLevel := Level{
			Level: levelNum,
			Nodes: []LevelNode{},
		}

		nextQueue := []string{}

		// 处理当前层级的所有节点
		for _, nodeName := range queue {
			processed[nodeName] = true

			// 添加到当前层级
			parentName := ""
			if node, exists := nodeMap[nodeName]; exists {
				parentName = node.ParentName
			}

			currentLevel.Nodes = append(currentLevel.Nodes, LevelNode{
				Name:       nodeName,
				ParentName: parentName,
			})

			// 处理子节点
			for _, child := range graph[nodeName] {
				inDegree[child]--
				if inDegree[child] == 0 {
					nextQueue = append(nextQueue, child)
				}
			}
		}

		if len(currentLevel.Nodes) > 0 {
			levels = append(levels, currentLevel)
		}

		queue = nextQueue
		levelNum++
	}

	// 检查是否所有节点都被处理（如果没有，说明有环）
	for nodeName := range allNodes {
		if !processed[nodeName] {
			return nil, errors.New("图中存在环，这不是一个有效的DAG")
		}
	}

	return levels, nil
}

// 示例使用
func main() {
	// 测试用例1：正常的DAG
	nodes1 := []Node{
		{Name: "B", ParentName: "A"},
		{Name: "C", ParentName: "A"},
		{Name: "D", ParentName: "B"},
		{Name: "E", ParentName: "C"},
		{Name: "F", ParentName: "D"},
		{Name: "F", ParentName: "E"}, // F有两个父节点
	}

	fmt.Println("测试用例1 - 正常DAG:")
	levels, err := detectCycleAndGetLevels(nodes1)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
	} else {
		for _, level := range levels {
			fmt.Printf("层级 %d:\n", level.Level)
			for _, node := range level.Nodes {
				fmt.Printf("  节点: %s, 父节点: %s\n", node.Name, node.ParentName)
			}
		}
	}

	// 测试用例2：有环的图
	nodes2 := []Node{
		{Name: "B", ParentName: "A"},
		{Name: "C", ParentName: "B"},
		{Name: "A", ParentName: "C"}, // 形成环
	}

	fmt.Println("测试用例2 - 有环的图:")
	levels, err = detectCycleAndGetLevels(nodes2)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
	} else {
		for _, level := range levels {
			fmt.Printf("层级 %d:\n", level.Level)
			for _, node := range level.Nodes {
				fmt.Printf("  节点: %s, 父节点: %s\n", node.Name, node.ParentName)
			}
		}
	}
}
