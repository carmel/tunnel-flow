package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/carmel/tunnel-flow/config"
	"github.com/carmel/tunnel-flow/tunnel"
	"github.com/carmel/tunnel-flow/util"
	"github.com/hashicorp/yamux"
)

func main() {

	cfg, err := config.LoadClient("client_config.yaml")
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	tlsConfig, err :=
		util.LoadTlsConfig("../server.crt", "../server.crt", "../server.key", false)
	if err != nil {
		log.Fatal("Failed to load tls config:", err)
	}

	conn, err := tls.Dial("tcp", cfg.Server, tlsConfig)
	if err != nil {
		log.Fatal(err)
	}

	session, err := tunnel.NewSession(cfg.ClientID, conn, false)
	if err != nil {
		panic(err)
	}

	// 认证成功后注册所有服务
	for _, svc := range cfg.Services {
		frame := &tunnel.ControlFrame{
			Type: tunnel.FrameRegister,
			Payload: map[string]string{
				"client_id": cfg.ClientID,
				"service":   svc.Name,
				"local":     svc.Local,
			},
		}
		session.SendFrame(frame)
	}

	frame := &tunnel.ControlFrame{
		Type: tunnel.FrameAuth,
		Payload: map[string]string{
			"client_id": "my-client-123",
			"token":     "secret-token",
		},
	}

	session.SendFrame(frame)

	stream, err := session.AcceptStream()
	if err != nil {
		panic(err)
	}
	err = frame.Read(stream)
	if err != nil {
		panic(err)
	}

	if frame.Type != tunnel.FrameAuthResult || frame.Payload["status"] != "ok" {
		panic("Authentication failed: " + frame.Payload["message"])
	}

	fmt.Println("✅ Auth successful:", frame.Payload["message"])

	// 认证之后
	frame.Type = tunnel.FrameRegister
	frame.Payload = map[string]string{
		"client_id": "my-client-123",
		"service":   "web",
		"local":     "127.0.0.1:8080",
	}

	err = session.SendFrame(frame)
	if err != nil {
		panic(err)
	}
	fmt.Println("✅ Registered service 'web' -> 127.0.0.1:8080")

	go session.ListenLoop(func(stream *yamux.Stream) {

		frame := &tunnel.ControlFrame{}
		err = frame.Read(stream)

		if frame.Type != tunnel.FrameNewStream {
			return
		}

		target := frame.Payload["target"]
		localConn, err := net.Dial("tcp", target)
		if err != nil {
			fmt.Println("❌ Dial failed:", err)
			return
		}

		go io.Copy(localConn, stream)
		go io.Copy(stream, localConn)
	})

	// 向服务端发送消息
	fmt.Fprintf(stream, "hello from client\n")

	// 读取响应
	buf := make([]byte, 1024)
	n, _ := stream.Read(buf)
	fmt.Println("Received:", string(buf[:n]))
}
