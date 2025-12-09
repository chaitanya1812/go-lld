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
	for i := 0; i < numFacts; i++ {
		// add to wait group before calling go routine
		wg.Add(1)
		go func() {
			// mark as done after the execution of go routine
			// should be triggered at the end, so first defer
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
			results <- string(bytes) + fmt.Sprintf("id : %d", i)
		}()
	}
	// wait for all the go routines to complete and close the results
	// using go routine becuase we can consume the avaialble results while this go-routine takes care about waiting and closing channel.
	go func() {
		wg.Wait()
		// triggers signal that says "no more values will be sent".
		close(results) // closing channel is mandatory to terminate the loop of consuming channel
	}()

	// consume the results when something is available in results
	for res := range results {
		fmt.Println(res)
	}

}
