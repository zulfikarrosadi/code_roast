package user

type User struct {
	Id        string
	Fullname  string
	Email     string
	Password  string
	CreatedAt int64
	Roles     []Roles
}

type Roles struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}
