package tunnel

import (
	"sync"
)

type SessionManager struct {
	sync.RWMutex
	sessions map[string]*Session
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*Session),
	}
}

func (m *SessionManager) Set(clientID string, session *Session) {
	m.Lock()
	defer m.Unlock()
	m.sessions[clientID] = session
}

// 如果你只关心 session 是否存在，不需要区分“key 不存在”和“值为 nil”，可以直接返回 *Session，代码更简洁。
// 但如果需要明确区分，保留 (*Session, bool) 更安全、健壮。
func (m *SessionManager) GetSession(clientID string) (*Session, bool) {
	m.RLock()
	defer m.RUnlock()
	sess, ok := m.sessions[clientID]
	return sess, ok
}
