package resultview

// ScrollState manages virtual scrolling over a row cache.
type ScrollState struct {
	TotalEstimate int // -1 if unknown
	LoadedRows    int
	PageSize      int
	CacheLimit    int
	AllLoaded     bool

	rows       [][]string
	fetchStart int // first logical row index in cache
}

// DefaultPageSize is the default number of rows per fetch.
const DefaultPageSize = 200

// DefaultCacheLimit is the max rows kept in memory.
const DefaultCacheLimit = 10000

// NewScrollState creates a ScrollState with defaults.
func NewScrollState() *ScrollState {
	return &ScrollState{
		TotalEstimate: -1,
		PageSize:      DefaultPageSize,
		CacheLimit:    DefaultCacheLimit,
	}
}

// Reset clears all cached data.
func (s *ScrollState) Reset() {
	s.rows = nil
	s.LoadedRows = 0
	s.TotalEstimate = -1
	s.AllLoaded = false
	s.fetchStart = 0
}

// SetRows replaces the cache with the given rows.
func (s *ScrollState) SetRows(rows [][]string) {
	s.rows = rows
	s.LoadedRows = len(rows)
	s.fetchStart = 0
	s.evict()
}

// AppendRows adds rows to the end of the cache, evicting oldest if over limit.
func (s *ScrollState) AppendRows(rows [][]string) {
	s.rows = append(s.rows, rows...)
	s.LoadedRows = len(s.rows)
	s.evict()
}

// Rows returns the cached rows.
func (s *ScrollState) Rows() [][]string {
	return s.rows
}

// NeedsFetch returns true if the viewport is approaching the edge of loaded data.
func (s *ScrollState) NeedsFetch(viewportEnd int) bool {
	if s.AllLoaded {
		return false
	}
	threshold := s.PageSize / 4
	if threshold < 1 {
		threshold = 1
	}
	return viewportEnd+threshold >= s.LoadedRows
}

func (s *ScrollState) evict() {
	if s.CacheLimit <= 0 || len(s.rows) <= s.CacheLimit {
		return
	}
	excess := len(s.rows) - s.CacheLimit
	s.rows = s.rows[excess:]
	s.fetchStart += excess
	s.LoadedRows = len(s.rows)
}
