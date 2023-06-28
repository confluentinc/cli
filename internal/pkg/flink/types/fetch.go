package types

type FetchState int32

const (
	Paused    FetchState = iota // auto fetch was Paused
	Completed                   // arrived at last page, fetch is Completed
	Failed                      // fetching next page Failed
	Running                     // auto fetch is Running
)
