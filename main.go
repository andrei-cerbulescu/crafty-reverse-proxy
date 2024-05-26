package main

import (
	"crypto/tls"
	"net/http"
	"sync"
)

func main() {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	config := getConfig()
	var wg sync.WaitGroup
	
	for _, address := range config.Addresses {
		wg.Add(1)
		go func(address ServerType){
			defer wg.Done()
			handleServer(address)
		}(address)
	}
	wg.Wait()
}
