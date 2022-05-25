package models

type Book struct {
	ID     string  `json:"id" validate:"omitempty,uuid"`
	Isbn   string  `json:"isbn" validate:"required,min=10,max=10"`
	Title  string  `json:"title" validate:"required"`
	Genre  string  `json:"genre" validate:"required"`
	Author *Author `json:"author" validate:"required"`
}

type Author struct {
	Name    string `json:"name" validate:"required"`
	Country string `json:"country" validate:"required"`
}

type User struct {
	UserId          string `json:"userid" validate:"omitempty"`
	FirstName       string `json:"firstname" validate:"required"`
	LastName        string `json:"lastname" validate:"required"`
	Email           string `json:"email" validate:"required"`
	Password        string `json:"password" validate:"required"`
	ConfirmPassword string `json:"confirmpassword" validate:"omitempty"`
}

type Response struct {
	Data       interface{}
	Status     string
	StatusCode int
}
