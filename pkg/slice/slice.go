package slice

// IntSliceContainsIntValue returns true if value appears in slice
// at least once or false otherwise. It returns false if slice is empty.
func IntSliceContainsIntValue(slice []int, value int) bool {
	for _, num := range slice {
		if num == value {
			return true
		}
	}
	return false
}

// IsUniqueIntSlice returns true if slice contains unique values and false otherwise.
// It returns true if slice is empty.
func IsUniqueIntSlice(slice []int) bool {
	var temp []int

	for _, num := range slice {
		if IntSliceContainsIntValue(temp, num) {
			return false
		}
		temp = append(temp, num)
	}

	return true
}
