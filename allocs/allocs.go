//go:build !solution

package allocs

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
)

// implement your Counter below
type allocCounter struct {
	words map[string]int
}

func (a *allocCounter) Count(r io.Reader) error {
	reader := bufio.NewReader(r)

	for {
		line, isPrefix, err := reader.ReadLine()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		if isPrefix {
			return nil
		}

		words := strings.Split(string(line), " ")
		for _, word := range words {
			a.words[word]++
		}
	}
}

func (a *allocCounter) String() string {
	keys := make([]string, 0, len(a.words))
	for key := range a.words {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	builder := &strings.Builder{}
	for _, key := range keys {
		line := fmt.Sprintf("word '%s' has %d occurrences\n", key, a.words[key])
		_, _ = builder.WriteString(line)
	}

	return builder.String()
}

func NewEnhancedCounter() Counter {
	return &allocCounter{words: make(map[string]int)}
}
