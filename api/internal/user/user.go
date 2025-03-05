package user

type User struct {
	Id        string  `json:"id"`
	Fullname  string  `json:"fullname"`
	Email     string  `json:"email"`
	Password  string  `json:"password"`
	CreatedAt int64   `json:"created_at"`
	Roles     []Roles `json:"roles"`
}

type Roles struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}
