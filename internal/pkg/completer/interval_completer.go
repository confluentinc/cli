package completer

//
//import (
//	"sync"
//	"time"
//
//	"github.com/c-bata/go-prompt"
//	"github.com/spf13/cobra"
//)
//
//const (
//	DefaultFetchInterval = 10 * time.Second
//)
//
//type CacheCompleter struct {
//	Completer
//
//	// map[string][]prompt.Suggest
//	cachedSuggestionsByCmd *sync.Map
//
//	// map[string]time.Time
//	lastFetchTimeByCmd *sync.Map
//
//	FetchInterval time.Duration
//}
//
//func NewCacheCompleter(completer Completer, fetchInterval time.Duration) *CacheCompleter {
//	c := &CacheCompleter{
//		Completer:                    completer,
//		cachedSuggestionsByCmd:       new(sync.Map),
//		lastFetchTimeByCmd:           new(sync.Map),
//		FetchInterval:                fetchInterval,
//	}
//	return c
//}
//
//func (c *CacheCompleter) Complete(d prompt.Document) []prompt.Suggest {
//
//	cachedSuggestions, ok := c.cachedSuggestionsByCmd.Load(key)
//	if !ok {
//		// TODO: Call completionFunc.
//		return []prompt.Suggest{}
//	}
//	if suggestions, ok := cachedSuggestions.([]prompt.Suggest); ok {
//		return suggestions
//	} else {
//		return []prompt.Suggest{}
//	}
//}
//
//// StartBackgroundUpdates starts fetching suggestions for all registered commands on a separate goroutine.
//func (c *CacheCompleter) StartBackgroundFetching() {
//	for range time.Tick(c.FetchInterval) {
//		c.updateAllSuggestions()
//	}
//}
//
//func (c *CacheCompleter) updateAllSuggestions() {
//	c.SuggestionFunctionsByCmdName.Range(func(key, sFunc interface{}) bool {
//		annotation := key.(string)                        // Will panic if not string.
//		suggestionFunc := sFunc.(func() []prompt.Suggest) // Will also cause panic if not the correct type.
//		c.updateSuggestion(annotation, suggestionFunc)
//		return true
//	})
//}
//
//func (c *CacheCompleter) updateSuggestion(annotation string, suggestionFunc func() []prompt.Suggest) {
//	c.lastFetchTimeByCmd.Store(annotation, time.Now())
//	suggestions := suggestionFunc()
//	c.cachedSuggestionsByCmd.Store(annotation, suggestions)
//}
