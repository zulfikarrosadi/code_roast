package user

type User struct {
	Id        string  `json:"id"`
	Fullname  string  `json:"fullname"`
	Email     string  `json:"email,omitempty"`
	Password  string  `json:"password,omitempty"`
	CreatedAt int64   `json:"created_at,omitempty"`
	Roles     []Roles `json:"roles,omitempty"`
}

type Roles struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}
