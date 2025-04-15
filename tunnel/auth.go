package tunnel

var allowedClients = map[string]string{
	"my-client-123": "secret-token",
}

func Authenticate(clientID, token string) bool {
	if expected, ok := allowedClients[clientID]; ok {
		return token == expected
	}
	return false
}
