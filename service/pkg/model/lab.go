package model

type UserType string

const (
	LABTYPE    = "lab"
	NORNAMTYPE = "normal-user"
)

type LabInfo struct {
	AccessKey         string   `json:"accessKey"`
	AccessSecret      string   `json:"accessSecret"`
	Name              string   `json:"name"`
	DisplayName       string   `json:"displayName"`
	SignupApplication string   `json:"signupApplication"`
	Avatar            string   `json:"avatar"`
	Owner             string   `json:"owner"`
	Type              UserType `json:"type"`
	Password          string   `json:"password"`
}

type LabInfoResp struct {
	Status string `json:"status"`
	Msg    string `json:"msg"`
	Sub    string `json:"sub"`
	Name   string `json:"name"`
}

type LabAkSk struct {
	AccessKey    string `json:"access_key"`
	AccessSecret string `json:"access_secret"`
}

type UserData struct {
	Owner             string `json:"owner"`
	Name              string `json:"name"`
	ID                string `json:"id"`
	Avatar            string `json:"avatar"`
	Type              string `json:"type"`
	DisplayName       string `json:"displayName"`
	SignupApplication string `json:"signupApplication"`
	AccessToken       string `json:"accessToken"`
	AccessKey         string `json:"accessKey"`
	AccessSecret      string `json:"accessSecret"`
	Phone             string `json:"phone"`
	Status            int    `json:"status"`
	UserNo            string `json:"user_no"`
	Email             string `json:"email"`
}

type UserInfo struct {
	Status string    `json:"status"`
	Msg    string    `json:"msg"`
	Sub    string    `json:"sub"`
	Name   string    `json:"name"`
	Data   *UserData `json:"data,omitempty"`
}
