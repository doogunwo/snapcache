package data

import (
    "crypto/rand"
    "encoding/hex"
)

// KeyValueGenerator는 랜덤 키-값 쌍을 생성하는 구조체입니다.
type KeyValueGenerator struct {
    keySize   int
    valueSize int
}

// NewKeyValueGenerator는 새로운 KeyValueGenerator를 생성합니다.
func NewKeyValueGenerator(keySize, valueSize int) *KeyValueGenerator {
    return &KeyValueGenerator{
        keySize:   keySize,
        valueSize: valueSize,
    }
}

// GenerateRandomString는 주어진 크기의 랜덤 문자열을 생성합니다.
func (gen *KeyValueGenerator) GenerateRandomString(size int) string {
    bytes := make([]byte, size)
    _, err := rand.Read(bytes)
    if err != nil {
        panic(err)
    }
    return hex.EncodeToString(bytes)[:size]
}

// GenerateKeyValuePair는 24바이트 크기의 키-값 쌍을 생성합니다.
func (gen *KeyValueGenerator) GenerateKeyValuePair() (string, string) {
    key := gen.GenerateRandomString(gen.keySize)
    value := gen.GenerateRandomString(gen.valueSize)
    return key, value
}


