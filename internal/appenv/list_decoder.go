package appenv

import "strings"

// decode env list values
//
// e.g: NAMES=[some name 1 , somename2 , etc...]
//
// NOTE: you can not have "," in the values. not supported.
func DecodeEnvList(listAsStr string) []string {
	if string(listAsStr[0]) != "[" || string(listAsStr[len(listAsStr)-1]) != "]" {
		return []string{}
	}
	listAsStr = listAsStr[1 : len(listAsStr)-1]

	arr := strings.Split(listAsStr, ",")
	if len(arr) == 0 {
		return []string{}
	}

	for i, v := range arr {
		arr[i] = strings.TrimSpace(v)
	}

	return arr
}
