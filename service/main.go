package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/panjf2000/ants/v2"
	"github.com/scienceol/studio/service/cmd/api"
	"github.com/scienceol/studio/service/cmd/schedule"
	"github.com/scienceol/studio/service/pkg/utils"
	"github.com/spf13/cobra"
)

func main() {
	rootCtx := utils.SetupSignalContext()
	root := &cobra.Command{
		SilenceUsage: true,
		Short:        "unilab",
		Long:         "unilab 智能实验室后端服务",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
	root.SetContext(rootCtx)
	root.AddCommand(api.NewWeb())
	root.AddCommand(api.NewMigrate())
	root.AddCommand(schedule.New())

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func main1() {
	fmt.Println("================start ")
	defer ants.Release()

	var wg sync.WaitGroup

	// 创建一个容量为10的协程池
	pool, _ := ants.NewPool(10)
	defer pool.Release()

	for _ = range 10 {
		fmt.Println("waiting:   ", pool.Waiting())

		wg.Add(1)
		pool.Submit(func() {
			defer wg.Done()
			fmt.Println("Hello from goroutine pool")
			time.Sleep(time.Second * 10)
		})

		fmt.Println("waiting:   ", pool.Waiting())
	}

	wg.Wait()
}
