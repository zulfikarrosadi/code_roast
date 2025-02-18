package user

type User struct {
	Id        string `json:"id"`
	Fullname  string `json:"fullname"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	CreatedAt int64  `json:"created_at"`
}

type Authentication struct {
	Id           string `json:"id"`
	RefreshToken string `json:"refresh_token"`
	LastLogin    int64  `json:"last_login"`
	RemoteIP     string `json:"remote_ip"`
	Agent        string `json:"agent"`
	UserId       string `json:"user_id"`
}

type userCreateRequest struct {
	Id        string `json:"id"`
	Fullname  string `json:"fullname"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	CreatedAt int64  `json:"created_at"`
	Authentication
}

type userCreateResponse struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Fullname string `json:"fullname"`
}

type authResponse struct {
	User         userCreateResponse `json:"user"`
	AccessToken  string             `json:"accessToken"`
	RefreshToken string             `json:"refreshToken"`
}

type userLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
