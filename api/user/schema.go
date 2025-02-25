package user

type User struct {
	Id        string `json:"id"`
	Fullname  string `json:"fullname"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	CreatedAt int64  `json:"created_at"`
}

// doesn't have json tag because the data is generated from system
type authentication struct {
	id           string
	refreshToken string
	lastLogin    int64
	remoteIP     string
	agent        string
	userId       string
}

type userCreateRequest struct {
	Id       string `json:"id"`
	Fullname string `json:"fullname"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Agent    string `json:"agent"`
	RemoteIp string `json:"remote_ip"`
}

type userCreateResponse struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Fullname string `json:"fullname"`
}

type authResponse struct {
	User         userCreateResponse `json:"user"`
	AccessToken  string             `json:"access_token"`
	RefreshToken string             `json:"refresh_token"`
}

type userLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
