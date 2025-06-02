package tracker

import (
	"os"
	"sync"
)

func StartSaver(path string, saveCh <-chan string, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic(err)
		}
		defer file.Close()

		for id := range saveCh {
			file.WriteString(id + "\n")
		}
	}()
}
