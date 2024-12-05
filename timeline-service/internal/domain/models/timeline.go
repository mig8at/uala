package models

type Tweet struct {
	ID       string `json:"id"`
	UserID   string `json:"userId"`
	Content  string `json:"content"`
	Likes    int    `json:"likes"`
	Shares   int    `json:"shares"`
	Comments int    `json:"comments"`
}

type User struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

type Timeline struct {
	ID       string `json:"id"`
	Content  string `json:"content"`
	Likes    int    `json:"likes"`
	Shares   int    `json:"shares"`
	Comments int    `json:"comments"`

	UserID   string `json:"userId"`
	Name     string `json:"name"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}
