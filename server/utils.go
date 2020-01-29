package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"reflect"

	"google.golang.org/api/androidpublisher/v3"
)

func getPackageNameFromArgs(arg string, userID string, packageList []PackageInfo, aliases map[string]string) (packageName string, ok bool) {
	if contains(packageList, PackageInfo{Name: arg, UserID: userID}) {
		return arg, true
	}
	packageName, ok = aliases[arg]
	return packageName, ok
}

func contains(slice []PackageInfo, value PackageInfo) bool {
	for _, elem := range slice {
		if elem.Name == value.Name &&
			elem.UserID == value.UserID {
			return true
		}
	}
	return false
}

func getAliasesForPackage(packageName string, aliases map[string]string) []string {
	result := []string{}
	for k, v := range aliases {
		if v == packageName {
			result = append(result, k)
		}
	}
	return result
}

func removeElement(list []*androidpublisher.Review, index int) []*androidpublisher.Review {
	var newList []*androidpublisher.Review
	if index == len(list)-1 {
		newList = list[:index]
	} else {
		newList = append(list[:index], list[index+1:]...)
	}
	return newList
}

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}

func isField(fieldName string, s interface{}) bool {
	valueS := reflect.ValueOf(s)
	field := valueS.FieldByName(fieldName)
	return field.IsValid()
}

func encrypt(key []byte, text string) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	msg := pad([]byte(text))
	ciphertext := make([]byte, aes.BlockSize+len(msg))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], msg)
	finalMsg := base64.URLEncoding.EncodeToString(ciphertext)
	return finalMsg, nil
}

func decrypt(key []byte, text string) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	decodedMsg, err := base64.URLEncoding.DecodeString(text)
	if err != nil {
		return "", err
	}

	if (len(decodedMsg) % aes.BlockSize) != 0 {
		return "", errors.New("blocksize must be multipe of decoded message length")
	}

	iv := decodedMsg[:aes.BlockSize]
	msg := decodedMsg[aes.BlockSize:]

	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(msg, msg)

	unpadMsg, err := unpad(msg)
	if err != nil {
		return "", err
	}

	return string(unpadMsg), nil
}

func pad(src []byte) []byte {
	padding := aes.BlockSize - len(src)%aes.BlockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padtext...)
}

func unpad(src []byte) ([]byte, error) {
	length := len(src)
	unpadding := int(src[length-1])

	if unpadding > length {
		return nil, errors.New("unpad error. This could happen when incorrect encryption key is used")
	}

	return src[:(length - unpadding)], nil
}
