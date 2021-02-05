package net

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/wufe/polo/models"
	"github.com/wufe/polo/services"

	log "github.com/sirupsen/logrus"
)

func findApplicationByName(applications *[]*models.Application, name string) *models.Application {
	var foundApplication *models.Application
	for _, application := range *applications {
		if strings.ToLower(application.Name) == strings.ToLower(name) {
			foundApplication = application
		}
	}
	return foundApplication
}

func (server *HTTPServer) serveManager(res http.ResponseWriter, req *http.Request) {
	if server.isDev {
		req.URL.Path = fmt.Sprintf("%s%s", StaticFolderPath, StaticManagerFile)
		server.serveReverseProxy(server.devServerURL, res, req, nil) // webpack dev server
	} else {

		file, err := (*server.fileSystem).Open(StaticManagerFile)
		content, err := ioutil.ReadAll(file)
		if err != nil {
			log.Errorf("Could not read " + StaticManagerFile)
			return
		}
		res.WriteHeader(200)
		res.Write(content)
	}
}

func (server *HTTPServer) getDashboard(
	res http.ResponseWriter,
	req *http.Request,
	ps httprouter.Params,
) {
	server.serveManager(res, req)
}

func (server *HTTPServer) getSessionStatus(
	res http.ResponseWriter,
	req *http.Request,
	ps httprouter.Params,
) {
	server.serveManager(res, req)
}

func (server *HTTPServer) getApplicationsAPI(
	res http.ResponseWriter,
	req *http.Request,
	ps httprouter.Params,
) {
	resString, resStatus := buildResponse(ResponseObjectWithResult{
		ResponseObject: ResponseObject{
			Message: "Ok",
		},
		Result: server.Configuration.Applications,
	}, 200)
	res.Header().Add("Content-Type", "application/json")
	res.WriteHeader(resStatus)
	res.Write(resString)
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

func (server *HTTPServer) getSessionByUUIDAPI(
	res http.ResponseWriter,
	req *http.Request,
	ps httprouter.Params,
) {
	uuid := ps.ByName("uuid")

	var foundSession *models.Session
	for _, session := range server.SessionHandler.GetAllAliveSessions() {
		if session.UUID == uuid {
			foundSession = session
		}
	}

	if foundSession == nil {
		resString, resStatus := buildResponse((ResponseObjectWithFailingReason{
			ResponseObject: ResponseObject{
				Message: "Not found",
			},
		}), 404)
		res.Header().Add("Content-Type", "application/json")
		res.WriteHeader(resStatus)
		res.Write(resString)
		return
	}

	resString, resStatus := buildResponse(ResponseObjectWithResult{
		ResponseObject: ResponseObject{
			Message: "Ok",
		},
		Result: foundSession,
	}, 200)
	res.Header().Add("Content-Type", "application/json")
	res.WriteHeader(resStatus)
	res.Write(resString)
}

func (server *HTTPServer) deleteSessionByUUIDAPI(
	res http.ResponseWriter,
	req *http.Request,
	ps httprouter.Params,
) {
	uuid := ps.ByName("uuid")

	var foundSession *models.Session
	for _, session := range server.SessionHandler.GetAllAliveSessions() {
		if session.UUID == uuid {
			foundSession = session
		}
	}

	if foundSession == nil {
		resString, resStatus := buildResponse((ResponseObjectWithFailingReason{
			ResponseObject: ResponseObject{
				Message: "Not found",
			},
		}), 404)
		res.Header().Add("Content-Type", "application/json")
		res.WriteHeader(resStatus)
		res.Write(resString)
		return
	}

	server.SessionHandler.DestroySession(foundSession)
	resString, resStatus := buildResponse(ResponseObject{
		Message: "Ok",
	}, 200)
	res.Header().Add("Content-Type", "application/json")
	res.WriteHeader(resStatus)
	res.Write(resString)
}

func (server *HTTPServer) getSessionLogsAndStatusByUUIDAPI(
	res http.ResponseWriter,
	req *http.Request,
	ps httprouter.Params,
) {
	uuid := ps.ByName("uuid")

	var foundSession *models.Session
	for _, session := range server.SessionHandler.GetAllAliveSessions() {
		if session.UUID == uuid {
			foundSession = session
		}
	}

	if foundSession == nil {
		resString, resStatus := buildResponse((ResponseObjectWithFailingReason{
			ResponseObject: ResponseObject{
				Message: "Not found",
			},
		}), 404)
		res.Header().Add("Content-Type", "application/json")
		res.WriteHeader(resStatus)
		res.Write(resString)
		return
	}

	logs := foundSession.Logs
	lastLogUUID := ps.ByName("last_log")

	if lastLogUUID != "" && lastLogUUID != "<none>" {
		logs = []models.Log{}
		afterLastLog := false
		for _, log := range foundSession.Logs {
			if afterLastLog {
				logs = append(logs, log)
			}
			if log.UUID == lastLogUUID {
				afterLastLog = true
			}
		}
	}

	resString, resStatus := buildResponse(ResponseObjectWithResult{
		ResponseObject: ResponseObject{
			Message: "Ok",
		},
		Result: struct {
			Logs   []models.Log         `json:"logs"`
			Status models.SessionStatus `json:"status"`
		}{
			Logs:   logs,
			Status: foundSession.Status,
		},
	}, 200)
	res.Header().Add("Content-Type", "application/json")
	res.WriteHeader(resStatus)
	res.Write(resString)
}

func (server *HTTPServer) getSessionAgeByUUIDAPI(
	res http.ResponseWriter,
	req *http.Request,
	ps httprouter.Params,
) {
	uuid := ps.ByName("uuid")

	var foundSession *models.Session
	for _, session := range server.SessionHandler.GetAllAliveSessions() {
		if session.UUID == uuid {
			foundSession = session
		}
	}

	if foundSession == nil {
		resString, resStatus := buildResponse((ResponseObjectWithFailingReason{
			ResponseObject: ResponseObject{
				Message: "Not found",
			},
		}), 404)
		res.Header().Add("Content-Type", "application/json")
		res.WriteHeader(resStatus)
		res.Write(resString)
		return
	}

	resString, resStatus := buildResponse(ResponseObjectWithResult{
		ResponseObject: ResponseObject{
			Message: "Ok",
		},
		Result: foundSession.MaxAge,
	}, 200)
	res.Header().Add("Content-Type", "application/json")
	res.WriteHeader(resStatus)
	res.Write(resString)
}

func (server *HTTPServer) postTrackSessionByUUIDAPI(
	res http.ResponseWriter,
	req *http.Request,
	ps httprouter.Params,
) {
	uuid := ps.ByName("uuid")

	var foundSession *models.Session
	for _, session := range server.SessionHandler.GetAllAliveSessions() {
		if session.UUID == uuid {
			foundSession = session
		}
	}

	if foundSession == nil {
		resString, resStatus := buildResponse((ResponseObjectWithFailingReason{
			ResponseObject: ResponseObject{
				Message: "Not found",
			},
		}), 404)
		res.Header().Add("Content-Type", "application/json")
		res.WriteHeader(resStatus)
		res.Write(resString)
		return
	}

	resString, resStatus := buildResponse(ResponseObjectWithResult{
		ResponseObject: ResponseObject{
			Message: "Ok",
		},
	}, 200)
	server.trackSession(res, foundSession)
	res.Header().Add("Content-Type", "application/json")
	res.WriteHeader(resStatus)
	res.Write(resString)
}

func (server *HTTPServer) postUntrackSessionAPI(
	res http.ResponseWriter,
	req *http.Request,
	ps httprouter.Params,
) {
	server.untrackSession(res)
	resString, resStatus := buildResponse(ResponseObjectWithResult{
		ResponseObject: ResponseObject{
			Message: "Ok",
		},
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
		Checkout        string `json:"checkout"`
		ApplicationName string `json:"applicationName"`
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

	// Looking for the required application
	foundApplication := findApplicationByName(&server.Configuration.Applications, sessionCreationInput.ApplicationName)
	if foundApplication == nil {
		resString, resStatus := buildResponse(ResponseObjectWithFailingReason{
			ResponseObject: ResponseObject{
				Message: "Bad request",
			},
			Reason: fmt.Sprintf("Application named %s not found", sessionCreationInput.ApplicationName),
		}, 404)
		res.Header().Add("Content-Type", "application/json")
		res.WriteHeader(resStatus)
		res.Write(resString)
		return
	}

	// Building the new session
	response := server.SessionHandler.RequestNewSession(&services.SessionBuildInput{
		Checkout:    sessionCreationInput.Checkout,
		Application: foundApplication,
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
		Result: response.Session,
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
