package mem

import "github.com/jonabc/test-repo/gss"

// WordService is the implementation of the gss WordService interface.
type WordService struct {
}

// NewWordService creates a new WordService instance.
func NewWordService() *WordService {
	return &WordService{}
}

// ReverseWord returns the given Word, with the Name field reversed.
func (w *WordService) ReverseWord(word *gss.Word) (*gss.Word, error) {
	return &gss.Word{
		Name: reverse(word.Name),
	}, nil
}

func reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}
