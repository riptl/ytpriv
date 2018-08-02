package net

import (
	"net/http"
	"os"
	"bufio"
	"encoding/base64"
	"io/ioutil"
	"encoding/json"
)

type DebugTransport struct{
	File *os.File
	Writer *bufio.Writer
}

type RequestObj struct{
	Host   string `json:"host"`
	Method string `json:"method"`
	Path   string `json:"path"`
	Header http.Header `json:"header"`
}

type ResponseObj struct{
	Kind string `json:"type"`
	StatusCode int `json:"status_code"`
	Status string `json:"status"`
	Header http.Header `json:"header"`
	Body string `json:"body"`
}

type LogObj struct{
	Request *RequestObj  `json:"request"`
	Err string `json:"error,omitempty"`
	Response *ResponseObj `json:"response,omitempty"`
}

func (t DebugTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	// Behave just like MainTransport
	resp, err = MainTransport{}.RoundTrip(req)

	// Initialize request info
	reqObj := RequestObj{req.Host, req.Method, req.RequestURI, req.Header }
	logObj := LogObj{ Request: &reqObj }

	if err != nil {
		logObj.Err = err.Error()
	} else {
		// Read whole body
		data, encErr := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if encErr != nil {
			logObj.Err = encErr.Error()
			goto logJson
		}
		// Replace r.Body with in-memory response
		req.Body = NewMemReadCloser(data)

		// Response to base64 for logging
		dataEnc := base64.StdEncoding.EncodeToString(data)

		logObj.Response = &ResponseObj{
			"success",
			resp.StatusCode,
			resp.Status,
			resp.Header,
			dataEnc,
		}
	}

	logJson:
	jsn, jsonErr := json.Marshal(logObj)
	if jsonErr != nil { panic(jsonErr) }

	t.Writer.Write(jsn)
	t.Writer.WriteByte('\n')

	return
}
