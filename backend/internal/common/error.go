// Package common
package common

type MyError struct {
	Data    any    `json:"data"`
	Status  int    `json:"status" default:"500"`
	Message string `json:"message" default:"message"`
}

func (e *MyError) Error() string {
	return e.Message
}

func (e *MyError) GetStatus() int {
	return e.Status
}
