package model

type HelloRequest struct {
	F1 string `json:"f1"`
	F2 string `json:"f2" validate:"required"`
}

type HelloResponse struct{}
