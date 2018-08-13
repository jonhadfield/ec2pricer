package ec2pricer

import "strings"

func StringInSlice(a string, list []string, caseInsensitive bool) bool {
	for _, b := range list {
		if caseInsensitive && strings.ToLower(b) == strings.ToLower(a) {
			return true
		} else if b == a {
			return true
		}
	}
	return false
}

func GetKeyByVal(input map[string]string, val string, caseInsensitive bool) string {
	for k, v := range input {
		if caseInsensitive && strings.ToLower(v) == strings.ToLower(val) {
			return k
		} else if v == val {
			return k
		}
	}
	return ""
}

func GetMatchingKey(input map[string]string, key string, caseInsensitive bool) string {
	for k := range input {
		if caseInsensitive && strings.ToLower(k) == strings.ToLower(key) {
			return k
		} else if k == key {
			return k
		}
	}
	return ""
}
