package event

import "fmt"

type EventType string

func (e EventType) String() string {
	return string(e)
}

const (
	Remove EventType = "remove"
	Write  EventType = "write"
	Create EventType = "create"
	Rename EventType = "rename"
)

type Event struct {
	Type     EventType
	FromPath string
	ToPath   string
}

func (e Event) String() string {
	s := fmt.Sprintf("rsult %s", e.FromPath)
	if e.Type == Rename {
		s += fmt.Sprintf(" -> %s", e.ToPath)
	}
	s += fmt.Sprintf(" [%s]", e.Type)
	return s
}
