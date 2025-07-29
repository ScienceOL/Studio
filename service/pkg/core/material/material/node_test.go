package material

import (
	"context"
	"fmt"
	"testing"

	"github.com/scienceol/studio/service/pkg/core/material"
)

func TestNodeLevel(t *testing.T) {
	nodes := []*material.Node{
		{
			Name:   "1",
			Parent: "2",
		},
		{
			Name:   "2",
			Parent: "3",
		},
		{
			Name:   "3",
			Parent: "5",
		},
		{
			Name:   "4",
			Parent: "5",
		},
		{
			Name:   "5",
			Parent: "6",
		},
		{
			Name:   "6",
			Parent: "",
		},
		{
			Name:   "7",
			Parent: "9",
		},
	}

	result := sortNodeLevel(context.Background(), nodes)
	fmt.Println(result)
}
