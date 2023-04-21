package data

import (
	"net/http"

	"github.com/google/uuid"
)

// SuccessFunc returns two bools:
//
//	bool	true == Success
//	bool	true == retry (only applies if success is false)
type SuccessFunc func(resp *http.Response) (bool, bool)

type HttpRequest struct {
	Id      string        `json:"id"`
	Req     *http.Request `json:"req"`
	Client  *http.Client  `json:"-"`
	Success []SuccessFunc `json:"-"`
}

func NewHttpRequest(id string, req *http.Request, successFunc ...SuccessFunc) *HttpRequest {
	idToUse := id
	if id == "" {
		idToUse = uuid.NewString()
	}
	return &HttpRequest{
		Id:      idToUse,
		Req:     req,
		Success: successFunc,
	}
}
