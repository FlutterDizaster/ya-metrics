package utils

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
)

// ReadPEM - читает PEM файл и возвращает блок для дальнейшего парсина.
// Чтобы получить сертивикат или приватный ключь, нужно использовать ReadCertificate или ReadPrivateKey.
// Возвращает ошибку, если не удалось прочитать файл или в файле нет блоков PEM.
func ReadPEM(path string) (*pem.Block, error) {
	// Получение ключа из файла
	certData, err := os.ReadFile(path)
	if err != nil {
		return nil, ErrReadFile
	}

	// Парсинг блока
	certBlock, _ := pem.Decode(certData)
	if certBlock == nil {
		return nil, errors.New("failed to parse certificate PEM")
	}

	return certBlock, nil
}

// ReadPrivateKey - читает приватный ключ из PEM файла.
// Возвращает ошибку, если не удалось прочитать файл или в файле нет блоков с приватным ключом.
func ReadPrivateKey(path string) (*rsa.PrivateKey, error) {
	keyPEM, err := ReadPEM(path)
	if err != nil {
		return nil, err
	}

	if keyPEM.Type != "RSA PRIVATE KEY" {
		return nil, errors.New("RSA private key not found")
	}

	return x509.ParsePKCS1PrivateKey(keyPEM.Bytes)
}

func ReadPublicKey(path string) (*rsa.PublicKey, error) {
	keyPEM, err := ReadPEM(path)
	if err != nil {
		return nil, err
	}

	if keyPEM.Type != "PUBLIC KEY" {
		return nil, errors.New("public key not found")
	}

	pubKey, err := x509.ParsePKIXPublicKey(keyPEM.Bytes)
	if err != nil {
		return nil, errors.New("failed to parse public key")
	}

	// Приведение к rsa.PublicKey
	pubKeyRSA, ok := pubKey.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("failed to convert public key to rsa")
	}

	return pubKeyRSA, nil
}
