package servicemanager

import "testing"

func TestPartition(t *testing.T) {

	testData := map[string][]string{
		"abcdef": []string{"ab", "cd", "ef"},
		"abcde":  []string{"ab", "cd", "e"},
		"ab":     []string{"ab"},
		"":       []string{},
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
