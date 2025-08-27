package model

type Handle struct {
	Label      string `json:"label"`
	DataKey    string `json:"data_key"`
	DataType   string `json:"data_type"`
	DataSource string `json:"data_source"`
	HandlerKey string `json:"handler_key"`
}

type ActionHandle struct {
	Input  []*Handle `json:"input"`
	Output []*Handle `json:"output"`
}
