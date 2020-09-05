package utils

import (
	"crypto/md5"
	"encoding/hex"
	"gopkg.in/gomail.v2"
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

//通过邮箱发送信息
func SendMail(mailTo []string, subject string, body string) error {
	// 设置邮箱主体
	mailConn := map[string]string{
		"user": "xxx@qq.com",  //发送人邮箱（邮箱以自己的为准）
		"pass": "xxx",         //发送人邮箱的密码，现在可能会需要邮箱 开启授权密码后在pass填写授权码
		"host": "smtp.qq.com", //邮箱服务器（此时用的是qq邮箱）
	}

	m := gomail.NewMessage(
		//发送文本时设置编码，防止乱码。 如果txt文本设置了之后还是乱码，那可以将原txt文本在保存时
		//就选择utf-8格式保存
		gomail.SetEncoding(gomail.Base64),
	)
	m.SetHeader("From", m.FormatAddress(mailConn["user"], "集合吧懒虫们")) // 添加别名
	m.SetHeader("To", mailTo...)                                     // 发送给用户(可以多个)
	m.SetHeader("Subject", subject)                                  // 设置邮件主题
	m.SetBody("text/html", body)                                     // 设置邮件正文

	/*
	   创建SMTP客户端，连接到远程的邮件服务器，需要指定服务器地址、端口号、用户名、密码，如果端口号为465的话，
	   自动开启SSL，这个时候需要指定TLSConfig
	*/
	d := gomail.NewDialer(mailConn["host"], 465, mailConn["user"], mailConn["pass"]) // 设置邮件正文
	//d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	err := d.DialAndSend(m)
	return err
}
