package service

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"testing"
	"time"
)

func TestPasswordEncrypt(t *testing.T) {
	password := []byte("123456a@")
	encrypted, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	assert.NoError(t, err)
	println(string(encrypted))
	err = bcrypt.CompareHashAndPassword(encrypted, []byte("123456a@"))
	assert.NoError(t, err)
}

func TestParseTime(t *testing.T) {
	birthday := "2017-12-05"
	dateFormat := "2006-01-02"

	birth, _ := time.ParseInLocation(dateFormat, birthday, time.Local)
	birthUnix := birth.UnixMilli()
	fmt.Println(birth)
	fmt.Println(birthUnix)

	birthTime := time.Unix(0, birthUnix*int64(time.Millisecond))
	// 将 time.Time 类型转换为字符串
	birthdaySting := birthTime.Format(dateFormat)

	fmt.Println(birthTime)
	fmt.Println(birthdaySting)

}
