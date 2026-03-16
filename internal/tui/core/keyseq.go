package core

// KeySeq tracks a two-key sequence (e.g., "gt", "dd", "dR").
type KeySeq struct {
	first string
}

// Start begins a new sequence with the given first key.
func (ks *KeySeq) Start(key string) { ks.first = key }

// Active returns true if a sequence is in progress.
func (ks *KeySeq) Active() bool { return ks.first != "" }

// Consume returns and clears the first key.
func (ks *KeySeq) Consume() string {
	f := ks.first
	ks.first = ""
	return f
}

// Clear cancels any pending sequence.
func (ks *KeySeq) Clear() { ks.first = "" }
