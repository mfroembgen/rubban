package autoindexpattern

import (
	"context"
	"fmt"
	"sync"

	"github.com/sherifabdlnaby/gpool"
	"github.com/sherifabdlnaby/rubban/rubban/kibana"
)

//Run Run Auto Index Pattern creation task
func (a *AutoIndexPattern) Run(ctx context.Context) {
	//// Set for Found Patterns ( a set datastructes using Map )
	newIndexPatterns := make(map[string]kibana.IndexPattern)

	// Send Requests Concurrently
	pool := gpool.NewPool(a.concurrency)
	wg := sync.WaitGroup{}
	for _, generalPattern := range a.GeneralPatterns {
		multipleGeneralPatterns := generalPattern
		mx := sync.Mutex{}
		wg.Add(1)
		err := pool.Enqueue(ctx, func() {
			defer wg.Done()
			indexPatterns := a.getIndexPattern(ctx, multipleGeneralPatterns)

			// Add Result to global Result
			mx.Lock()
			for _, pattern := range indexPatterns {
				newIndexPatterns[pattern.Title] = pattern
			}
			mx.Unlock()
		})

		if err != nil {
			wg.Done()
		}
	}

	// Wait for all above jobs to Return
	wg.Wait()
	pool.Stop()

	// Create List from The Map Set we create
	indexPatterns := make([]kibana.IndexPattern, 0)
	for _, pattern := range newIndexPatterns {
		indexPatterns = append(indexPatterns, pattern)
	}

	err := a.kibana.BulkCreateIndexPattern(ctx, indexPatterns)
	if err != nil {
		a.log.Errorw("Failed to bulk create new index patterns", "error", err.Error())
	}

	a.log.Infow(fmt.Sprintf("Successfully created %d Index Patterns.", len(newIndexPatterns)), "Index Patterns", newIndexPatterns)

}

//Name Return Task Name
func (a *AutoIndexPattern) Name() string {
	return a.name
}
