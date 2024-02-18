package util

import (
	"math/rand"
	"strings"
)

func RandomInt(min, max int64) int64 {
	return min + rand.Int63n(max-min+1)
}

const alphabet = "abcdefghijklmnopqrstuvwxyz"

func RandomString(n int) string {
	var sb strings.Builder
	k := len(alphabet)
	for i := 0; i < k; i++ {
		ch := alphabet[rand.Intn(k)]
		sb.WriteByte(ch)
	}
	return sb.String()
}

func RandomOwner() string {
	return RandomString(6)
}

func RandomMoney() int64 {
	return RandomInt(0, 1000)
}

func RandomCurrency() string {

	currency := []string{"usd", "eur", "cad"}
	k := len(currency)
	return currency[rand.Intn(k)]
}
