package filenode

import "time"

type DateType int

const (
	MILLI DateType = iota
	NANO
)

type MetaData struct {
	IsDir      bool   `json:"is_dir"`
	Sum        string `json:"sum"`
	Size       int64  `json:"size"`
	CreatedAt  int64  `json:"created_at"`
	Permission string `json:"permission"`
}

func (m MetaData) CreatedDate(t DateType) time.Time {
	switch t {
	case NANO:
		return time.Unix(0, time.Unix(m.CreatedAt, 0).UnixNano()).UTC()
	default:
		return time.Unix(m.CreatedAt, 0).UTC()
	}
}

type ExtraPayload struct {
	UUID       string
	IsDir      bool
	Sum        string
	Size       int64
	CreatedAt  int64
	Permission string
}
