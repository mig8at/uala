package models

type Tweet struct {
	ID      string `json:"id"`
	UserID  string `json:"userId"`
	Content string `json:"content"`
	Likes   int    `json:"likes"`
	Shares  int    `json:"shares"`
}

type User struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

type Timeline struct {
	User  *User  `json:"user"`
	Tweet *Tweet `json:"tweet"`
}
