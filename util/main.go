package util

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

func StringUUID() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")

}

func IntUUID() uint32 {
	return uuid.New().ID()
}

func RandomString(l int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyz"
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for range l {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}

func LoadTlsConfig(caFile, certFile, keyFile string, insecure bool) (*tls.Config, error) {
	var (
		err  error
		ca   []byte
		cert tls.Certificate
	)
	ca, err = os.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load ca: %v", err)
	}

	cert, err = tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load cert: %v", err)
	}

	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(ca)

	// Create a TLS configuration with the loaded certificate
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      pool,
		// NextProtos:
		InsecureSkipVerify: insecure,
	}, nil

}
