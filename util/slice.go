package util

func IntSliceContains(slice []int, n int) bool {
	for _, v := range slice {
		if v == n {
			return true
		}
	}
	return false
}

func StringSliceContains(slice []string, text string) bool {
	for _, v := range slice {
		if v == text {
			return true
		}
	}
	return false
}
