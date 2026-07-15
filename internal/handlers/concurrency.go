package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func Concurrency(w http.ResponseWriter, r *http.Request) {
	// 1. Create a channel to communicate between goroutines
	// We make it buffered to prevent blocking if the receiver isn't ready immediately,
	// though for this small example unbuffered would work too if we read carefully.
	resultsChan := make(chan string, 5)

	// 2. Define some tasks to do
	tasks := []string{"task1", "task2", "task3", "task4", "task5"}

	// 3. Launch a goroutine for each task
	for _, task := range tasks {
		go func(t string) {
			// Simulate some work
			// In a real app, this could be an DB call, API request, calculation, etc.
			// sending the result to the channel
			result := fmt.Sprintf("%s processed", t)
			resultsChan <- result
		}(task)
	}

	// 4. Collect results
	var results []string
	// We expect exactly len(tasks) results
	for i := 0; i < len(tasks); i++ {
		res := <-resultsChan
		results = append(results, res)
	}
	// Ideally you would close the channel if you were ranging over it,
	// but since we count exactly, we don't strictly need to close it here
	// for control flow, but it's good practice when done sending.
	close(resultsChan)

	// Return response
	jsonResp, err := json.Marshal(map[string]interface{}{
		"results": results,
		"count":   len(results),
	})
	if err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResp)
}
