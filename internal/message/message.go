package message

type Status int

const (
	Error   Status = iota
	Warning Status = iota
	Info    Status = iota
	Debug   Status = iota
)

type Message struct {
	Status   Status
	Text     string
	Source string
}
