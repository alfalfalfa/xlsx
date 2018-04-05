package xlsx

import "runtime"

var AnonymousQueue chan func()

func init() {
	//worker
	AnonymousQueue = make(chan func())

	//consumer
	go func() {
		concurrent := runtime.NumCPU()
		for i := 0; i < concurrent; i++ {
			go func() {
				for task := range AnonymousQueue {
					if task == nil {
						return
					}
					task()
				}
			}()
		}
	}()
}
