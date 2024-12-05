package dto

type Tweet struct {
	ID      string `json:"id"`
	UserID  string `json:"userId" validate:"required,uuid"`
	Content string `json:"content" validate:"required,min=1,max=280"`
}

type CreateTweet struct {
	UserID  string   `json:"userId" validate:"required,uuid"`
	Content string   `json:"content" validate:"required,min=1,max=280"`
	Tags    []string `json:"tags" validate:"max=5,dive,min=5,max=20"`
}
