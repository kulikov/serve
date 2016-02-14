package utils

func Contains(elm string, list []string) bool {
	for _, v := range list {
		if v == elm {
			return true
		}
	}
	return false
}
