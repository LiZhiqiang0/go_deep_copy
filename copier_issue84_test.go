package copier_test

import (
	"time"
)

type Embedded struct {
	Field1 string
	Field2 string
}

type Embedder struct {
	Embedded
	PtrField *string
}

type Timestamps struct {
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type NotWork struct {
	ID      string  `json:"id"`
	UserID  *string `json:"user_id"`
	Name    string  `json:"name"`
	Website *string `json:"website"`
	Timestamps
}

type Work struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	UserID  *string `json:"user_id"`
	Website *string `json:"website"`
	Timestamps
}
