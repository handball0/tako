package md5

import (
	"crypto/md5"
	"encoding/hex"
)

// Md5Encode md5加密
func Md5Encode(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	cipherStr := h.Sum(nil)
	return hex.EncodeToString(cipherStr)
}

// Md5Password 密码加密
func Md5Password(pwd string, slat string) string {
	return Md5Encode(pwd + slat)
}

// ValidatePassword 验证密码
func ValidatePassword(pwd string, slat string, cipher string) bool {
	return Md5Password(pwd, slat) == cipher
}
