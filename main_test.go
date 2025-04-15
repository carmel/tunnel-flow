package tunnelflow

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"testing"
	"time"
)

// 生成临时证书
func generateTempCert() (tls.Certificate, error) {
	// 生成私钥
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, err
	}

	// 创建证书模板
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return tls.Certificate{}, err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"临时组织"},
			CommonName:   "localhost",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour), // 24小时有效期
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
		DNSNames:              []string{"localhost"},
	}

	// 创建自签名证书
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return tls.Certificate{}, err
	}

	// 编码为PEM格式
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})

	// 加载证书
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return tls.Certificate{}, err
	}

	return cert, nil
}

// 创建TLS服务器
func createTLSServer(address string) error {
	cert, err := generateTempCert()
	if err != nil {
		return fmt.Errorf("生成临时证书失败: %v", err)
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	listener, err := tls.Listen("tcp", address, config)
	if err != nil {
		return fmt.Errorf("监听失败: %v", err)
	}
	defer listener.Close()

	fmt.Printf("TLS服务器运行在 %s\n", address)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("接受连接失败: %v\n", err)
			continue
		}

		go handleConnection(conn)
	}
}

// 处理连接
func handleConnection(conn net.Conn) {
	defer conn.Close()

	fmt.Printf("接受来自 %s 的连接\n", conn.RemoteAddr())

	// 这里处理连接数据
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Printf("读取数据失败: %v\n", err)
		return
	}

	fmt.Printf("收到数据: %s\n", buffer[:n])
	conn.Write([]byte("已收到数据"))
}

// 创建TLS客户端
func createTLSClient(address string) error {
	// 忽略证书验证的配置
	config := &tls.Config{
		RootCAs: x509.NewCertPool(),
	}

	// 加载临时证书
	cert, err := generateTempCert()
	if err != nil {
		return fmt.Errorf("生成临时证书失败: %v", err)
	}

	// 将临时证书添加到RootCAs
	certPool := config.RootCAs
	certPool.AppendCertsFromPEM(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Certificate[0]}))

	conn, err := tls.Dial("tcp", address, config)
	if err != nil {
		return fmt.Errorf("连接失败: %v", err)
	}
	defer conn.Close()

	fmt.Printf("已连接到 %s\n", address)

	// 发送数据
	_, err = conn.Write([]byte("你好，服务器"))
	if err != nil {
		return fmt.Errorf("发送数据失败: %v", err)
	}

	// 接收响应
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		return fmt.Errorf("读取响应失败: %v", err)
	}

	fmt.Printf("服务器响应: %s\n", buffer[:n])
	return nil
}

func TestTls(t *testing.T) {

	// 可以根据需要选择运行服务器或客户端
	// 运行服务器
	go func() {
		err := createTLSServer(":8443")
		if err != nil {
			fmt.Printf("服务器错误: %v\n", err)
		}
	}()

	// 等待服务器启动
	time.Sleep(1 * time.Second)

	// 运行客户端
	err := createTLSClient("localhost:8443")
	if err != nil {
		fmt.Printf("客户端错误: %v\n", err)
	}
}
