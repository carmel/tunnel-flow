package tunnel

import (
	"fmt"
	"log"
	"net"
	"sync/atomic"

	"github.com/hashicorp/yamux"
)

type Session struct {
	id       string
	conn     net.Conn
	mux      *yamux.Session
	closed   atomic.Bool
	stopChan chan struct{}
}

func NewSession(id string, conn net.Conn, isServer bool) (*Session, error) {
	cfg := yamux.DefaultConfig()
	cfg.EnableKeepAlive = true
	cfg.AcceptBacklog = 128

	var mux *yamux.Session
	var err error
	if isServer {
		mux, err = yamux.Server(conn, cfg)
	} else {
		mux, err = yamux.Client(conn, cfg)
	}
	if err != nil {
		return nil, err
	}

	s := &Session{
		id:       id,
		conn:     conn,
		mux:      mux,
		stopChan: make(chan struct{}),
	}
	return s, nil
}

func (s *Session) ID() string {
	return s.id
}

func (s *Session) Conn() net.Conn {
	return s.conn
}

// func (s *Session) Mux() *yamux.Session {
// 	return s.mux
// }

func (s *Session) Close() {
	if s.closed.CompareAndSwap(false, true) {
		close(s.stopChan)
		s.mux.Close()
		s.conn.Close()
	}
}

func (s *Session) IsClosed() bool {
	return s.closed.Load()
}

func (s *Session) Done() <-chan struct{} {
	return s.stopChan
}

// 用于客户端开启数据/控制流
// func (s *Session) OpenStream() (net.Conn, error) {
// 	return s.mux.OpenStream()
// }

// 用于服务端接收流（控制帧 / 数据流复用）
func (s *Session) AcceptStream() (net.Conn, error) {
	return s.mux.AcceptStream()
}

// 📤 打开控制流发送控制帧
// func (ts *Session) SendFrame(data []byte) error {
// 	stream, err := ts.mux.OpenStream()
// 	if err != nil {
// 		return err
// 	}
// 	defer stream.Close()

// 	_, err = stream.Write(data)
// 	return err
// }

func (ts *Session) SendFrame(frame *ControlFrame) error {
	stream, err := ts.mux.OpenStream()
	if err != nil {
		return fmt.Errorf("failed to open control stream: %w", err)
	}
	defer stream.Close()

	// var data []byte
	// data, err = frame.Encode()
	// if err != nil {
	// 	return fmt.Errorf("failed to encode control frame: %w", err)
	// }
	// _, err = stream.Write(data)

	return frame.Write(stream)
}

// 📥 等待服务端指令：控制流监听（服务端）
func (ts *Session) ListenLoop(handle func(*yamux.Stream)) {
	for {
		stream, err := ts.mux.AcceptStream()
		if err != nil {
			if ts.IsClosed() {
				return
			}
			log.Println("control stream accept error:", err)
			continue
		}

		go func(s *yamux.Stream) {
			defer s.Close()
			handle(s)
		}(stream)
	}
}
