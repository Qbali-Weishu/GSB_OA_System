package utils

import (
	"regexp"
	"time"
	"unicode/utf8"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
var phoneRegex = regexp.MustCompile(`^1[3-9]\d{9}$`)

// ValidateEmail 验证邮箱格式
func ValidateEmail(email string) bool {
	return emailRegex.MatchString(email)
}

// ValidatePhone 验证手机号格式（中国大陆11位手机号）
func ValidatePhone(phone string) bool {
	return phoneRegex.MatchString(phone)
}

// ValidateDateRange 验证开始日期不晚于结束日期，且开始日期不早于今天
func ValidateDateRange(start, end time.Time) error {
	today := Today()
	if start.Before(today) {
		return BadRequest("开始日期不能早于今天", nil)
	}
	if end.Before(start) {
		return BadRequest("结束日期不能早于开始日期", nil)
	}
	// 单次请假最长不超过 30 个自然日
	if int(end.Sub(start).Hours()/24) > 30 {
		return BadRequest("单次请假不能超过30个自然日", nil)
	}
	return nil
}

// ValidateStringLength 验证字符串长度（按 UTF-8 字符数计）
func ValidateStringLength(s string, min, max int) bool {
	n := utf8.RuneCountInString(s)
	return n >= min && n <= max
}
