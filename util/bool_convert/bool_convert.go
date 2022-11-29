package boolconvert

import "errors"

func BoolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func Atob(s string) (bool, error) {
	if s == "true" {
		return true, nil
	} else if s == "false" {
		return false, nil
	}

	return false, errors.New("invalid parameter: " + s)
}

func BoolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
