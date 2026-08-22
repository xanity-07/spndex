package errs

// type AppErrors interface {
// 	error
// 	HTTPStatus() int
// 	ErrCode() string
// }

// type ValidationError struct {
// 	Code    string       `json:"code"`
// 	Message string       `json:"message"`
// 	Errors  []FieldError `json:"errors"`
// 	Status  int          `json:"status"`
// }

// func (e *ValidationError) Error() string {
// 	return e.Message
// }

// func (e *ValidationError) HTTPStatus() int {
// 	return e.Status
// }

// func (e *ValidationError) ErrCode() string {
// 	return e.Code
// }

// type UnauthorizedError struct {
// 	Code    string  `json:"code"`
// 	Message string  `json:"message"`
// 	Status  int     `json:"status"`
// 	Action  *Action `json:"action"`
// }

// func (e *UnauthorizedError) Error() string {
// 	return e.Message
// }

// func (e *UnauthorizedError) HTTPStatus() int {
// 	return e.Status
// }

// func (e *UnauthorizedError) ErrCode() string {
// 	return e.Code
// }
