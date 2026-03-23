package data

// import (
// 	"awesomeProject/dataBase"
// 	"awesomeProject/logger"
// 	"awesomeProject/node"
// 	"fmt"
// )

// var (
// 	InitDone = false
// )

// // TODO init
// func RunAll() {
// 	if InitDone {
// 		fmt.Printf("[Data] RunAll was done Before!\n")
// 		return
// 	}
// 	dataBase.Db.ResetDB()
// 	f := logger.InitLogFileWithName("create.log")
// 	defer f.Close()
// 	MakeLG(base)
// 	// plugin.MakeLG()
// 	InitDone = true
// }

// func MakeLG(group map[string]TR) {
// 	for _, p := range group {
// 		node.CreateLG(p.tv, p.rv)
// 	}
// }

// // func MakeLG(group map[string]TR) {
// // 	// 1. 控制并发数（根据CPU核心数或数据库连接池大小调整）
// // 	const maxWorkers = 16
// // 	sem := make(chan struct{}, maxWorkers)

// // 	// 2. 等待组，用于等待所有goroutine完成
// // 	var wg sync.WaitGroup

// // 	// 3. 遍历group，启动goroutine处理每个p
// // 	for _, p := range group {
// // 		// 捕获循环变量（避免goroutine共享同一变量）
// // 		pair := p
// // 		wg.Add(1)
// // 		sem <- struct{}{} // 占用一个并发槽位

// // 		go func() {
// // 			defer func() {
// // 				wg.Done() // 标记goroutine完成
// // 				<-sem     // 释放并发槽位
// // 				// 可选：捕获panic，避免单个goroutine崩溃导致整个程序退出
// // 				if r := recover(); r != nil {
// // 					fmt.Printf("处理数据时发生错误: %v, tv:%+v, rv:%+v\n", r, pair.tv, pair.rv)
// // 				}
// // 			}()

// // 			// 执行核心逻辑
// // 			node.CreateLG(pair.tv, pair.rv)
// // 		}()
// // 	}

// // 	// 4. 等待所有goroutine完成
// // 	wg.Wait()
// // 	close(sem)
// // }
