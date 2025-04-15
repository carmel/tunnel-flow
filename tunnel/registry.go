package tunnel

import (
	"sync"
)

type ServiceMapping struct {
	ClientID string
	Local    string
}

var serviceRegistry = struct {
	sync.RWMutex
	services map[string]ServiceMapping
}{
	services: make(map[string]ServiceMapping),
}

func RegisterService(name, clientID, local string) {
	serviceRegistry.Lock()
	defer serviceRegistry.Unlock()
	serviceRegistry.services[name] = ServiceMapping{
		ClientID: clientID,
		Local:    local,
	}
}

func GetService(name string) (ServiceMapping, bool) {
	serviceRegistry.RLock()
	defer serviceRegistry.RUnlock()
	svc, ok := serviceRegistry.services[name]
	return svc, ok
}
