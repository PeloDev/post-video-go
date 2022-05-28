package utils

func Contains(arr []string, search string) bool {
	for _, item := range arr {
		if item == search {
			return true
		}
	}
	return false
}
