package servicemanager

import "testing"

func TestPartition(t *testing.T) {

	testData := map[string][]string{
		"abcdef": {"ab", "cd", "ef"},
		"abcde":  {"ab", "cd", "e"},
		"ab":     {"ab"},
		"":       {},
	}

	for input, output := range testData {
		res := partition(input, 2)
		for i, r := range res {
			if output[i] != r {
				t.Errorf("input: [%s]:%d, expected [%s] got [%s]", input, i, output[i], r)
			}
		}
	}
}
