package util

import (
	"crypto/md5"
	"fmt"
	"strings"
)

// HashObject 序列化内容进行md5
func HashObject(data []byte) string {
	has := md5.Sum(data)
	return fmt.Sprintf("%x", has)
}

func Md5slice(clusters []string) string {
	if len(clusters) == 0 {
		return ""
	}
	str := strings.Join(clusters, "")
	data := []byte(str)
	has := md5.Sum(data)
	md5str := fmt.Sprintf("%x", has)
	return md5str
}
