package util

import (
	"reflect"
)

// Helper function to find if a elem in a slice
func HasElem(s interface{}, elem interface{}) bool {
	arrv := reflect.ValueOf(s)
	if arrv.Kind() == reflect.Slice {
		for i := 0; i < arrv.Len(); i++ {
			if arrv.Index(i).Interface() == elem {
				return true
			}
		}
	}
	return false
}

func SubtractStringSlice(slice, s []string) []string {
	for _, i := range s {
		for pos, j := range slice {
			if i == j {
				slice = append(slice[:pos], slice[pos+1:]...)
				break
			}
		}
	}
	return slice
}
