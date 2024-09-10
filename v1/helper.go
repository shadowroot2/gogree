package v1

func ArrayFlip(m []string) map[string]int {
	n := make(map[string]int, len(m))
	for i, v := range m {
		n[v] = i
	}
	return n
}
