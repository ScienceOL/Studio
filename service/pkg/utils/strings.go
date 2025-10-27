package utils

import "unicode"

func CamelToSnake(s string) string {
	var result []rune
	var prevLower bool // 记录前一个字符是否为小写

	for i, r := range s {
		if unicode.IsUpper(r) {
			// 不是首字符且前一个字符是小写，或者前一个字符存在且是大写但下一个字符是小写
			if i > 0 && (prevLower || (i+1 < len(s) && unicode.IsLower(rune(s[i+1])))) {
				result = append(result, '_')
			}
			result = append(result, unicode.ToLower(r))
			prevLower = false
		} else {
			result = append(result, r)
			prevLower = unicode.IsLower(r)
		}
	}

	return string(result)
}
