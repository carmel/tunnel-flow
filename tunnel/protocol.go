package tunnel

import (
	"encoding/json"
	"io"
)

const (
	// ...已有的类型
	FrameRegister FrameType = "REGISTER"
)

type FrameType string

const (
	FramePing       FrameType = "PING"
	FramePong       FrameType = "PONG"
	FrameNewStream  FrameType = "NEW_STREAM"
	FrameClose      FrameType = "CLOSE"
	FrameError      FrameType = "ERROR"
	FrameAuth       FrameType = "AUTH"
	FrameAuthResult FrameType = "AUTH_RESULT"
)

type ControlFrame struct {
	Type    FrameType         `json:"type"`
	Stream  string            `json:"stream,omitempty"`  // 可选：逻辑流 ID
	Payload map[string]string `json:"payload,omitempty"` // 任意键值对
}

func (frame *ControlFrame) Read(r io.Reader) error {
	buf := make([]byte, 4096)
	n, err := r.Read(buf)
	if err != nil {
		return err
	}

	err = json.Unmarshal(buf[:n], &frame)
	if err != nil {
		return err
	}
	return nil
}

func (frame *ControlFrame) Write(w io.Writer) error {
	data, err := json.Marshal(frame)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	_, err = w.Write(data)
	return err
}

// func (frame *ControlFrame) Encode() ([]byte, error) {
// 	return json.Marshal(frame)
// }

// func (frame *ControlFrame) Decode(data []byte) error {
// 	err := json.Unmarshal(data, &frame)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }
