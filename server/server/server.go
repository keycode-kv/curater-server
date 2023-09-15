package server

import (
	"sync"
)

func Start() (err error) {

	wg := new(sync.WaitGroup)

	wg.Add(1)

	go func() {
		startHTTPServer()
		wg.Done()
	}()

	wg.Wait()
	return
}
