package common

type Error struct {
	Msg string `json:"msg"`
}

type Resp struct {
	Code  int   `json:"code"`
	Error Error `json:"error"`
	Data  any   `json:"data"`
}
