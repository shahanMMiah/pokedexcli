package internal

import (
	"fmt"
	"testing"
	"time"
)

func TestCacheReap(t *testing.T) {
	cases := []struct {
		input    *Cache
		expected int
	}{
		{
			input:    NewCache(500 * time.Millisecond),
			expected: 0,
		},
	}
	fmt.Println("testing cache reeping")
	for _, c := range cases {

		dataVar := make([]byte, 100)
		c.input.Add("testURL", dataVar)
		time.Sleep(500 * time.Millisecond)

		if len(c.input.CacheMap) != c.expected {
			t.Errorf("%v reeping has failed", c.input.CacheMap)
		}
	}
}
