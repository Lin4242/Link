package transport

const (
	TypeMessage   = "msg"
	TypeDelivered = "delivered"
	TypeTyping    = "typing"
	TypeRead      = "read"
	TypeOnline    = "online"
	TypeOffline   = "offline"
	TypeError     = "error"
)

type Message struct {
	Type    string      `json:"t"`
	Payload interface{} `json:"p,omitempty"`
}
