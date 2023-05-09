package queue

import "strings"

// NormaliseQueueName is an encapsulated way of normalising all of
// the queue names.
//
// The idea is that every time a queue name is referenced/used, it is done via this
// function.  This ensures consistency at every point-of-use.
func NormaliseQueueName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}
