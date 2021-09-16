package response_builder

import (
	"encoding/json"
	"net/http"

	"github.com/wufe/polo/pkg/logging"
)

type ResponseBuilder struct {
	log logging.Logger
}

func NewResponseBuilder(logger logging.Logger) *ResponseBuilder {
	return &ResponseBuilder{log: logger}
}

type ResponseObject struct {
	Message string `json:"message"`
}

type ResponseObjectWithResult struct {
	ResponseObject
	Result interface{} `json:"result,omitempty"`
}

type ResponseObjectWithFailingReason struct {
	ResponseObject
	Reason interface{} `json:"reason,omitempty"`
}

func (b *ResponseBuilder) NotFound() ([]byte, int) {
	return b.BuildResponse(ResponseObjectWithFailingReason{
		ResponseObject{"Not found"},
		"Not found",
	}, 404)
}

func (b *ResponseBuilder) Ok(obj interface{}) ([]byte, int) {
	return b.BuildResponse(ResponseObjectWithResult{
		ResponseObject{"Ok"},
		obj,
	}, 200)
}

func (b *ResponseBuilder) ServerError(reason interface{}) ([]byte, int) {
	return b.BuildResponse(ResponseObjectWithFailingReason{
		ResponseObject{"Internal server error"},
		reason,
	}, 500)
}

func (b *ResponseBuilder) BadRequest() ([]byte, int) {
	return b.BuildResponse(ResponseObject{"Bad request"}, 400)
}

func (b *ResponseBuilder) BuildResponse(response interface{}, status int) ([]byte, int) {
	responseString, err := json.Marshal(response)
	if err != nil {
		b.log.Errorln("Could not serialize response object", err)
		return []byte(`{"message": "Internal server error"}`), 500
	} else {
		return responseString, status
	}
}

func (b *ResponseBuilder) OkOrNotFound(obj interface{}, status int) ([]byte, int) {

	var c []byte
	var s int

	if obj != nil {
		c, s = b.Ok(obj)
	} else {
		c, s = b.NotFound()
	}

	return c, s
}

func (b *ResponseBuilder) Write(w http.ResponseWriter) func(c []byte, s int) {
	return func(c []byte, s int) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(s)
		w.Write(c)
	}
}
