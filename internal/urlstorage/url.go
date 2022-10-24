package urlstorage

import (
	"crypto/sha512"
	"encoding/hex"
	"math/rand"
)

type ShortURL struct {
	url    []byte
	offset int
}

func NewShortURL(url string) *ShortURL {
	hash := sha512.New()
	hash.Write([]byte(url))
	return &ShortURL{
		url:    hash.Sum(nil),
		offset: 0,
	}
}

func (u *ShortURL) Next(size int) string {
	if u.offset+size > len(u.url) {
		u.url[rand.Intn(size-1)/2]++
		return hex.EncodeToString(u.url)[:size]
	}
	str := hex.EncodeToString(u.url)[u.offset : u.offset+size]
	u.offset++
	return str
}
