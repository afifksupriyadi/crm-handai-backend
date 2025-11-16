package response

// Response is a wrapper for http response
type Response struct {
	Status int         `json:"status"`
	Body   interface{} `json:"body"`
}

// Body is a wrapper for response body
type Body struct {
	Data    interface{} `json:"data"`
	Code    string      `json:"code"`
	Message string      `json:"message"`
}

// HumaError implements the huma.StatusError interface.
type HumaError struct {
	Body
	Status int `json:"-"`
}

// Error implements the error interface.
// It returns the error message from HumaError.Body
func (r *HumaError) Error() string {
	return r.Message
}

// GetStatus implements the huma.StatusError interface.
// It returns the status code from HumaError
func (r *HumaError) GetStatus() int {
	return r.Status
}
