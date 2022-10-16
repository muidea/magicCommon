package util

import (
	"crypto/md5"
)

func CryptoByMd5(data []byte) []byte {
	md5Ctx := md5.New()
	md5Ctx.Write(data)
	return md5Ctx.Sum(nil)
}
