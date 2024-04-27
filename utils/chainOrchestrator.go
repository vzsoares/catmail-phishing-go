package utils

func ChainOrchestrator(fn func() int) int {

	return fn()
}
