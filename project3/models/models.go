package models

type Book struct {
	ID     string  `json:"id" validate:"omitempty,uuid"`
	Isbn   string  `json:"isbn" validate:"required,min=10,max=10"`
	Title  string  `json:"title" validate:"required"`
	Genre  string  `json:"genre" validate:"required"`
	Author *Author `json:"author" validate:"required"`
}

type Author struct {
	Firstname string `json:"firstname" validate:"required"`
	Lastname  string `json:"lastname" validate:"required"`
	Age       int    `json:"age" validate:"required,min=1,max=100"`
}

type Response struct {
	Data       interface{}
	Status     string
	StatusCode int
}
