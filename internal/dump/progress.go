package dump

import (
	"bufio"
	"io"
	"regexp"
)

// ProgressEvent represents a single progress update from pg_dump.
type ProgressEvent struct {
	Object string // table name currently being dumped
	Index  int    // 1-based count of objects dumped so far
	Total  int    // expected total objects (0 if unknown)
	Done   bool   // true when dump is complete
	Err    error  // non-nil on error
}

var tableRe = regexp.MustCompile(`pg_dump: dumping contents of table "([^"]+)"`)

// ParseProgress reads pg_dump verbose stderr and emits ProgressEvents.
// total is the expected number of objects (used for percentage); pass 0 if unknown.
// The returned channel is closed after a Done or Err event.
func ParseProgress(r io.Reader, total int) <-chan ProgressEvent {
	ch := make(chan ProgressEvent)
	go func() {
		defer close(ch)
		scanner := bufio.NewScanner(r)
		index := 0
		for scanner.Scan() {
			line := scanner.Text()
			m := tableRe.FindStringSubmatch(line)
			if m == nil {
				continue
			}
			index++
			ch <- ProgressEvent{
				Object: m[1],
				Index:  index,
				Total:  total,
			}
		}
		if err := scanner.Err(); err != nil {
			ch <- ProgressEvent{Err: err}
			return
		}
		ch <- ProgressEvent{Done: true, Index: index, Total: total}
	}()
	return ch
}
