package tunnel

import (
	"fmt"
	"net"

	"github.com/carmel/tunnel-flow/util"
)

func StartForwarder(listenAddr, serviceName string, sessionManager *SessionManager) error {
	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return err
	}
	fmt.Printf("🌐 Listening on %s for service '%s'\n", listenAddr, serviceName)

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				fmt.Println("accept error:", err)
				continue
			}

			go handleForwardConn(conn, serviceName, sessionManager)
		}
	}()
	return nil
}

func handleForwardConn(conn net.Conn, serviceName string, sessionManager *SessionManager) {

	mapping, ok := GetService(serviceName)
	if !ok {
		fmt.Println("❌ No such service:", serviceName)
		return
	}

	session, found := sessionManager.GetSession(mapping.ClientID)
	if !found {
		fmt.Println("❌ No session for client:", mapping.ClientID)
		return
	}

	// stream, err := session.OpenStream()
	// if err != nil {
	// 	fmt.Println("❌ OpenStream failed:", err)
	// 	return
	// }

	// 发送 NEW_STREAM 控制帧，请求客户端连接其内网服务
	frame := &ControlFrame{
		Type:   FrameNewStream,
		Stream: "stream-" + util.RandomString(8), // 可选 stream ID
		Payload: map[string]string{
			"target": mapping.Local,
		},
	}
	err := session.SendFrame(frame)
	if err != nil {
		fmt.Println("❌ SendFrame failed:", err)
	}
}
