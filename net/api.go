package net

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/wufe/polo/models"
	"github.com/wufe/polo/services"

	log "github.com/sirupsen/logrus"
)

func (server *HTTPServer) getDashboard(
	res http.ResponseWriter,
	req *http.Request,
	ps httprouter.Params,
) {
	isDev := true
	if isDev {
		req.URL.Path = "/_polo_/static/dashboard.html"
		server.serveReverseProxy("http://localhost:9000/", res, req) // webpack dev server
	} else {
		path := filepath.Join(StaticFolder, "dashboard.html")
		content, err := ioutil.ReadFile(path)
		if err != nil {
			log.Errorf("Could not read %s", path)
		}
		res.WriteHeader(200)
		res.Write(content)
	}
}

func (server *HTTPServer) getSessionStatus(
	res http.ResponseWriter,
	req *http.Request,
	ps httprouter.Params,
) {
	res.WriteHeader(200)
	res.Write([]byte(fmt.Sprintf("Session status: %s", ps.ByName("uuid"))))
}

func (server *HTTPServer) getServicesAPI(
	res http.ResponseWriter,
	req *http.Request,
	ps httprouter.Params,
) {
	services, err := json.Marshal(server.Configuration.Services)
	if err != nil {
		log.Fatalf("Error serializing services", err)
	}
	res.Header().Add("Content-Type", "application/json")
	res.WriteHeader(200)
	res.Write(services)
}

func (server *HTTPServer) getAllSessionsAPI(
	res http.ResponseWriter,
	req *http.Request,
	ps httprouter.Params,
) {
	// content, err := json.Marshal(server.SessionHandler.GetAllSessions())
	resString, resStatus := buildResponse(ResponseObjectWithResult{
		ResponseObject: ResponseObject{
			Message: "Ok",
		},
		Result: server.SessionHandler.GetAllAliveSessions(),
	}, 200)
	res.Header().Add("Content-Type", "application/json")
	res.WriteHeader(resStatus)
	res.Write(resString)
}

func (server *HTTPServer) postSessionAPI(
	res http.ResponseWriter,
	req *http.Request,
	ps httprouter.Params,
) {

	// Decoding body
	sessionCreationInput := &struct {
		Checkout    string `json:"checkout"`
		ServiceName string `json:"serviceName"`
	}{}
	err := json.NewDecoder(req.Body).Decode(sessionCreationInput)
	if err != nil {
		resString, resStatus := buildResponse(ResponseObjectWithFailingReason{
			ResponseObject: ResponseObject{
				Message: "Bad request",
			},
			Reason: err.Error(),
		}, 400)
		res.Header().Add("Content-Type", "application/json")
		res.WriteHeader(resStatus)
		res.Write(resString)
		return
	}

	// Looking for the required service
	var foundService *models.Service
	for _, service := range server.Configuration.Services {
		if strings.ToLower(service.Name) == strings.ToLower(sessionCreationInput.ServiceName) {
			foundService = service
		}
	}
	if foundService == nil {
		resString, resStatus := buildResponse(ResponseObjectWithFailingReason{
			ResponseObject: ResponseObject{
				Message: "Bad request",
			},
			Reason: fmt.Sprintf("Service named %s not found", sessionCreationInput.ServiceName),
		}, 404)
		res.Header().Add("Content-Type", "application/json")
		res.WriteHeader(resStatus)
		res.Write(resString)
		return
	}

	// Building the new session
	response := server.SessionHandler.RequestNewSession(&services.SessionBuildInput{
		Checkout: sessionCreationInput.Checkout,
		Service:  foundService,
	})
	if response.Result == services.SessionBuildResultFailed {

		responseObject := ResponseObjectWithFailingReason{
			ResponseObject: ResponseObject{
				Message: "Internal server error",
			},
		}
		if response.FailingReason != "" {
			responseObject.Reason = response.FailingReason
		}

		resString, status := buildResponse(responseObject, 500)
		res.Header().Add("Content-Type", "application/json")
		res.WriteHeader(status)
		res.Write(resString)
		return

	}

	responseObject := ResponseObjectWithResult{
		ResponseObject: ResponseObject{
			Message: "Ok",
		},
		Result: response.Session.UUID,
	}

	resString, status := buildResponse(responseObject, 200)
	res.Header().Add("Content-Type", "application/json")
	res.WriteHeader(status)
	res.Write(resString)
}

func buildResponse(response interface{}, status int) ([]byte, int) {
	responseString, err := json.Marshal(response)
	if err != nil {
		log.Errorln("Could not serialize response object", err)
		return []byte(`{"message": "Internal server error"}`), 500
	} else {
		return responseString, status
	}
}

type ResponseObject struct {
	Message string `json:"message"`
}

type ResponseObjectWithResult struct {
	ResponseObject
	Result interface{} `json:"result"`
}

type ResponseObjectWithFailingReason struct {
	ResponseObject
	Reason interface{} `json:"reason,omitempty"`
}
