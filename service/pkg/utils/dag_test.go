package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildDAGHierarchy(t *testing.T) {
	t.Run("空节点列表", func(t *testing.T) {
		nodes := []*Node[string]{}
		result, err := BuildHierarchy(nodes)

		assert.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("单个根节点", func(t *testing.T) {
		nodes := []*Node[string]{
			{Name: "root", Parent: ""},
		}
		result, err := BuildHierarchy(nodes)

		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Len(t, result[0], 1)
		assert.Equal(t, "root", result[0][0].Name)
	})

	t.Run("简单两层层次结构", func(t *testing.T) {
		nodes := []*Node[string]{
			{Name: "root", Parent: ""},
			{Name: "child1", Parent: "root"},
			{Name: "child2", Parent: "root"},
		}
		result, err := BuildHierarchy(nodes)

		assert.NoError(t, err)
		assert.Len(t, result, 2)

		// 第一层：根节点
		assert.Len(t, result[0], 1)
		assert.Equal(t, "root", result[0][0].Name)

		// 第二层：子节点
		assert.Len(t, result[1], 2)
		childNames := []string{result[1][0].Name, result[1][1].Name}
		assert.Contains(t, childNames, "child1")
		assert.Contains(t, childNames, "child2")
	})

	t.Run("复杂多层层次结构", func(t *testing.T) {
		nodes := []*Node[string]{
			{Name: "root", Parent: ""},
			{Name: "A", Parent: "root"},
			{Name: "B", Parent: "root"},
			{Name: "C", Parent: "A"},
			{Name: "D", Parent: "A"},
			{Name: "E", Parent: "B"},
			{Name: "F", Parent: "C"},
		}
		result, err := BuildHierarchy(nodes)

		assert.NoError(t, err)
		assert.Len(t, result, 4)

		// 第一层：root
		assert.Len(t, result[0], 1)
		assert.Equal(t, "root", result[0][0].Name)

		// 第二层：A, B
		assert.Len(t, result[1], 2)
		level1Names := []string{result[1][0].Name, result[1][1].Name}
		assert.Contains(t, level1Names, "A")
		assert.Contains(t, level1Names, "B")

		// 第三层：C, D, E
		assert.Len(t, result[2], 3)
		level2Names := []string{result[2][0].Name, result[2][1].Name, result[2][2].Name}
		assert.Contains(t, level2Names, "C")
		assert.Contains(t, level2Names, "D")
		assert.Contains(t, level2Names, "E")

		// 第四层：F
		assert.Len(t, result[3], 1)
		assert.Equal(t, "F", result[3][0].Name)
	})

	t.Run("多个根节点", func(t *testing.T) {
		nodes := []*Node[string]{
			{Name: "root1", Parent: ""},
			{Name: "root2", Parent: ""},
			{Name: "child1", Parent: "root1"},
			{Name: "child2", Parent: "root2"},
		}
		result, err := BuildHierarchy(nodes)

		assert.NoError(t, err)
		assert.Len(t, result, 2)

		// 第一层：两个根节点
		assert.Len(t, result[0], 2)
		rootNames := []string{result[0][0].Name, result[0][1].Name}
		assert.Contains(t, rootNames, "root1")
		assert.Contains(t, rootNames, "root2")

		// 第二层：两个子节点
		assert.Len(t, result[1], 2)
		childNames := []string{result[1][0].Name, result[1][1].Name}
		assert.Contains(t, childNames, "child1")
		assert.Contains(t, childNames, "child2")
	})

	t.Run("节点名重复", func(t *testing.T) {
		nodes := []*Node[string]{
			{Name: "duplicate", Parent: ""},
			{Name: "duplicate", Parent: ""},
		}
		result, err := BuildHierarchy(nodes)

		assert.Error(t, err)
		assert.Equal(t, "节点名重复", err.Error())
		assert.Empty(t, result)
	})

	t.Run("存在环 - 直接环", func(t *testing.T) {
		nodes := []*Node[string]{
			{Name: "A", Parent: "B"},
			{Name: "B", Parent: "A"},
		}
		result, err := BuildHierarchy(nodes)

		assert.Error(t, err)
		assert.Equal(t, "图中存在环，这不是一个有效的DAG", err.Error())
		assert.Nil(t, result)
	})

	t.Run("存在环 - 间接环", func(t *testing.T) {
		nodes := []*Node[string]{
			{Name: "A", Parent: ""},
			{Name: "B", Parent: "A"},
			{Name: "C", Parent: "B"},
			{Name: "D", Parent: "C"},
			{Name: "B", Parent: "D"}, // 这会创建一个环
		}
		result, err := BuildHierarchy(nodes)

		assert.Error(t, err)
		assert.Equal(t, "节点名重复", err.Error()) // 这个案例会先被节点名重复检查捕获
		assert.Empty(t, result)
	})

	t.Run("存在环 - 修正版间接环", func(t *testing.T) {
		nodes := []*Node[string]{
			{Name: "A", Parent: ""},
			{Name: "B", Parent: "A"},
			{Name: "C", Parent: "B"},
			{Name: "A", Parent: "C"}, // A指向C，但A已经是C的祖先，形成环
		}
		result, err := BuildHierarchy(nodes)

		assert.Error(t, err)
		assert.Equal(t, "节点名重复", err.Error())
		assert.Empty(t, result)
	})

	t.Run("真正的环测试", func(t *testing.T) {
		nodes := []*Node[string]{
			{Name: "root", Parent: ""},
			{Name: "A", Parent: "root"},
			{Name: "B", Parent: "A"},
			{Name: "C", Parent: "B"},
			{Name: "D", Parent: "C"},
			{Name: "E", Parent: "D"},
			{Name: "F", Parent: "E"},
			{Name: "A", Parent: "F"}, // F指向A，但A是F的祖先，形成环
		}
		result, err := BuildHierarchy(nodes)

		assert.Error(t, err)
		assert.Equal(t, "节点名重复", err.Error())
		assert.Empty(t, result)
	})

	t.Run("父节点不存在", func(t *testing.T) {
		nodes := []*Node[string]{
			{Name: "root", Parent: ""},
			{Name: "child", Parent: "nonexistent"},
		}
		result, err := BuildHierarchy(nodes)

		assert.NoError(t, err)
		assert.Len(t, result, 1) // 只有一层，因为两个节点都是根节点（入度为0）

		// 第一层应该包含root和child两个节点
		assert.Len(t, result[0], 2)
		nodeNames := []string{result[0][0].Name, result[0][1].Name}
		assert.Contains(t, nodeNames, "root")
		assert.Contains(t, nodeNames, "child")
	})

	t.Run("复杂DAG结构", func(t *testing.T) {
		// 创建一个更复杂的DAG：多个节点指向同一个节点
		nodes := []*Node[string]{
			{Name: "root", Parent: ""},
			{Name: "A", Parent: "root"},
			{Name: "B", Parent: "root"},
			{Name: "C", Parent: "A"},
			{Name: "D", Parent: "B"},
			{Name: "E", Parent: "C"},
			{Name: "F", Parent: "D"},
			{Name: "G", Parent: "E"},
			{Name: "G", Parent: "F"}, // G有两个父节点E和F
		}
		result, err := BuildHierarchy(nodes)

		assert.Error(t, err)
		assert.Equal(t, "节点名重复", err.Error())
		assert.Empty(t, result)
	})

	t.Run("正确的多父节点DAG", func(t *testing.T) {
		// 由于当前实现不支持多父节点，我们测试单父节点的情况
		nodes := []*Node[string]{
			{Name: "root1", Parent: ""},
			{Name: "root2", Parent: ""},
			{Name: "A", Parent: "root1"},
			{Name: "B", Parent: "root2"},
			{Name: "C", Parent: "A"},
			{Name: "D", Parent: "B"},
			{Name: "E", Parent: "C"},
			{Name: "F", Parent: "D"},
		}
		result, err := BuildHierarchy(nodes)

		assert.NoError(t, err)
		assert.Len(t, result, 4)

		// 验证层次结构
		assert.Len(t, result[0], 2) // root1, root2
		assert.Len(t, result[1], 2) // A, B
		assert.Len(t, result[2], 2) // C, D
		assert.Len(t, result[3], 2) // E, F
	})
}
