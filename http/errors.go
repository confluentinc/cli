package http

import "fmt"

type apiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type ApiError struct {
	OrgId   int       `json:"organization_id"`
	UserId  int       `json:"user_id"`
	Err     *apiError `json:"error"`
}

func (e *ApiError) Error() string {
	return fmt.Sprintf("confluent (%v): %v (org:%v, user:%v)", e.Err.Code, e.Err.Message, e.OrgId, e.UserId)
}

func (e *ApiError) OrNil() error {
	if e.Err != nil {
		return e
	}
	return nil
}
