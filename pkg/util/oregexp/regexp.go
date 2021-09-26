package oregexp

import "regexp"

//
// RegexpReplace
// @Description:
// @param reg
// @param src
// @param temp
// @return string
//
func RegexpReplace(reg, src, temp string) string {
	result := []byte{}
	pattern := regexp.MustCompile(reg)
	for _, submatches := range pattern.FindAllStringSubmatchIndex(src, -1) {
		result = pattern.ExpandString(result, temp, src, submatches)
	}
	return string(result)
}
