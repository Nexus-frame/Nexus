package Nexus

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"path"
	"time"
)

func assert1(guard bool, text string) {
	if !guard {
		panic(text)
	}
}

func lastChar(str string) uint8 {
	if str == "" {
		panic("The length of the string can't be 0")
	}
	return str[len(str)-1]
}

func joinPaths(absolutePath, relativePath string) string {
	if relativePath == "" {
		return absolutePath
	}

	finalPath := path.Join(absolutePath, relativePath)
	if lastChar(relativePath) == '/' && lastChar(finalPath) != '/' {
		return finalPath + "/"
	}
	return finalPath
}

func GenerateUniqueString() string {
	date := time.Now().Format("20060102150405")
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		return fmt.Sprintf("%s%d", date, time.Now().UnixNano())
	}
	return fmt.Sprintf("%s%s", date, hex.EncodeToString(b))
}
