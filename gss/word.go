package gss

// WordService holds utilities related to word manipulation
type WordService interface {
	ReverseWord(w *Word) (*Word, error)
}

// Word is an example struct to demonstrate dependency passing
type Word struct {
	// Example field to include instead Word struct.
	Name string
}
