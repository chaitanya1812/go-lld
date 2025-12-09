package goconc

import (
	"fmt"
	"io"
	"net/http"
	"sync"
)

// The "Scatter-Gather" (WaitGroups)
// Concept: You need to fetch data from multiple sources in parallel and aggregate the results.

func Scatter_gather() {
	url := "https://catfact.ninja/fact"
	var wg sync.WaitGroup
	numFacts := 10
	results := make(chan string, numFacts)
	for range numFacts {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp, err := http.Get(url)
			if err != nil {
				results <- fmt.Sprintf("Unable to find the fact %s", err.Error())
				return
			}
			defer resp.Body.Close()
			bytes, err := io.ReadAll(resp.Body)
			if err != nil {
				results <- fmt.Sprintf("Unable to find the fact %s", err.Error())
				return
			}
			results <- string(bytes)
		}()
	}
	go func(){
		wg.Wait()
		close(results)
	}()

	for res := range results{
		fmt.Println(res)
	}

}
