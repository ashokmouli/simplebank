package util

import "testing"

func TestCheckPasswordMatch(t *testing.T) {
    clear := RandomString(10)
    hashed, _ := HashPassword(clear)

    err := CheckPassword(clear, hashed)

    if err != nil {
        t.Errorf("Expected nil, but got error: %v", err)
    }
}

func TestCheckPasswordEmptyHash(t *testing.T) {
    clear := RandomString(10)
    hash := ""

    err := CheckPassword(clear, hash)

    if err == nil {
        t.Error("Expected error, but got nil")
    }
}