package filenode

type MetaData struct {
	IsDir      bool   `json:"is_dir"`
	Sum        string `json:"sum"`
	Size       int64  `json:"size"`
	CreatedAt  int64  `json:"created_at"`
	Permission string `json:"permission"`
}

type ExtraPayload struct {
	UUID       string
	IsDir      bool
	Sum        string
	Size       int64
	CreatedAt  int64
	Permission string
}
