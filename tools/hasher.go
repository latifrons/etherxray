package tools

import "golang.org/x/crypto/sha3"

func Sha3HashString(input string) []byte {
	hasher := sha3.New256()
	hasher.Write([]byte(input))
	return hasher.Sum(nil)
}
