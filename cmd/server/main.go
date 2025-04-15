package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"

	"github.com/carmel/tunnel-flow/tunnel"
	"github.com/carmel/tunnel-flow/util"
	"github.com/hashicorp/yamux"
)

func main() {

	config, err :=
		util.LoadTlsConfig("../server.crt", "../server.crt", "../server.key", false)
	if err != nil {
		log.Fatal("Failed to load tls config:", err)
	}

	ln, err := tls.Listen("tcp", ":9000", config)
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	sm := tunnel.NewSessionManager()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Accept failed:", err)
			continue
		}
		go handleConnection(sm, conn)
	}

}

// func handleConnection(conn net.Conn) {
// 	session, err := tunnel.NewYamuxServer(conn)
// 	if err != nil {
// 		fmt.Println("Yamux error:", err)
// 		return
// 	}

// 	for {
// 		stream, err := session.AcceptStream()
// 		if err != nil {
// 			break
// 		}
// 		go handleStream(stream)
// 	}
// }

func handleConnection(sm *tunnel.SessionManager, conn net.Conn) {
	// session, err := tunnel.NewYamuxServer(conn)
	session, err := tunnel.NewSession(util.StringUUID(), conn, true)
	if err != nil {
		fmt.Println("create session error:", err)
		return
	}

	sm.Set(session.ID(), session)

	stream, err := session.AcceptStream()
	if err != nil {
		fmt.Println("AcceptStream error:", err)
		return
	}

	// 先进行认证
	frame := &tunnel.ControlFrame{}

	err = frame.Read(stream)
	if err != nil || frame.Type != tunnel.FrameAuth {
		fmt.Println("Auth failed: bad frame")
		stream.Close()
		return
	}

	clientID := frame.Payload["client_id"]
	token := frame.Payload["token"]

	authOK := tunnel.Authenticate(clientID, token)
	if !authOK {
		frame.Payload = map[string]string{
			"status":  "fail",
			"message": "unauthorized",
		}
		frame.Write(stream)
		stream.Close()
		return
	}

	// 回复认证成功
	frame.Payload = map[string]string{
		"status":  "ok",
		"message": "welcome " + clientID,
	}
	frame.Write(stream)

	// 后续可以记录 clientID -> session 映射
	go session.ListenLoop(func(s *yamux.Stream) {

		frame := &tunnel.ControlFrame{}
		err = frame.Read(stream)

		if err != nil || frame.Type != tunnel.FrameRegister {
			return
		}

		service := frame.Payload["service"]
		local := frame.Payload["local"]

		tunnel.RegisterService(service, clientID, local)

		fmt.Printf("✅ Client %s registered service '%s' -> %s\n", clientID, service, local)
	})
}
