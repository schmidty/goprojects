package main

var freqMap = make(map[string]int, 0)

func countLettersRecursive(s string) string {
	if len(s) == 0 {
		return s
	}
	freqMap[string(s[0])]++
	return countLettersRecursive(s[1:])
}

func countLettersLoop(s string) {
	for _, v := range s {
		freqMap[string(v)]++
	}
}
