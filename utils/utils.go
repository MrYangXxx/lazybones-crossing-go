package utils

import (
	"crypto/md5"
	"encoding/hex"
	rd "math/rand"
	"time"
)

/*
	生成随机数，纳秒时间戳作为种子
*/
func GetRandomNumber(length int) string {
	base := "0123456789"
	random := rd.New(rd.NewSource(time.Now().UnixNano()))

	result := ""
	for i := 0; i < length; i++ {
		index := random.Intn(len(base))
		result += string(base[index])
	}
	return result
}

/*
	生成随机数，根据传入的种子
*/
func GetRandomNumberBySource(length int, source int64) string {
	base := "0123456789"
	random := rd.New(rd.NewSource(source))

	result := ""
	for i := 0; i < length; i++ {
		index := random.Intn(len(base))
		result += string(base[index])
	}
	return result
}

//生成随机字符
func GetRandomString(length int) string {
	base := "abcdefghjkmnpqrstuvwxyzABCDEFGHJKMNPQRSTUVWXYZ123456789"
	random := rd.New(rd.NewSource(time.Now().UnixNano()))

	result := ""
	for i := 0; i < length; i++ {
		index := random.Intn(len(base))
		result += string(base[index])
	}
	return result
}

//MD5加密
func MD5(salt, value string) string {
	m5 := md5.New()
	m5.Write([]byte(value))
	m5.Write([]byte(salt))
	st := m5.Sum(nil)
	return hex.EncodeToString(st)
}

//array中是否包含val，存在返回其索引，不存在返回-1
func StringsContains(array []string, val string) (index int) {
	index = -1
	for i := 0; i < len(array); i++ {
		if array[i] == val {
			index = i
			return
		}
	}
	return
}
