package wcontext

import "errors"

var (
	QUERY_NOT_FOUND  = errors.New("query param cant't be found!")
	PARAMS_NOT_FOUND = errors.New("param cant't be found!")
	BODY_NOT_FOUND   = errors.New("body param cant't be found!")
)
