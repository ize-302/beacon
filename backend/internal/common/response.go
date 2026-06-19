// Package common
package common

type BaseResponseBody[T any] struct {
	Body struct {
		Data    T      `json:"data"`
		Status  bool   `json:"status"`
		Message string `json:"message"`
	}
}
