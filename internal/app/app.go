package app

import (
	"craftyreverseproxy/config"
	"crypto/tls"
	"net/http"
	"sync"
)

func Run() {
	var wg sync.WaitGroup

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	for _, address := range config.GetConfig().Addresses {
		wg.Add(1)
		go func(address config.ServerType) {
			defer wg.Done()
			handleServer(address)
		}(address)
	}
	wg.Wait()
}
