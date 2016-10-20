// Web Service request router
// a minimal API gateway implementation

// Author: kadirayk

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

type JsonRoot struct {
	Router Router
}

type Router struct {
	Port     string
	Handlers []Handler
}

type Handler struct {
	ListenPath         string
	HeaderName         string
	DefaultForwardPath DefaultForwardPath
	ForwardPaths       []ForwardPath
}

type DefaultForwardPath struct {
	Path        string
	ContentType string
	BasicAuth   BasicAuth
}

type ForwardPath struct {
	Condition   string
	Path        string
	ContentType string
	BasicAuth   BasicAuth
}

type BasicAuth struct {
	Username string
	Password string
}

func HttpHandler(w http.ResponseWriter, r *http.Request, router Router) {
	originalPath := r.URL.Path
	var handler Handler

	for _, v := range router.Handlers {
		if v.ListenPath == originalPath {
			handler = v
		}
	}
	if handler.ListenPath == "" {
		fmt.Printf("path not defined\n")
		return
	}

	headerValue := r.Header.Get(handler.HeaderName)

	var forwardPath ForwardPath

	var defaultForwardPath DefaultForwardPath

	var req *http.Request

	for _, v := range handler.ForwardPaths {
		if headerValue == v.Condition {
			forwardPath = v
			req = convertRequest(r, forwardPath)
			fmt.Printf("forwarded to:%v\n", forwardPath.Path)
			break
		}
	}

	if req == nil {
		defaultForwardPath = handler.DefaultForwardPath
		req = convertRequestWithDefaultPath(r, defaultForwardPath)
		fmt.Printf("forwarded to default :%v\n", defaultForwardPath.Path)
	}

	client := &http.Client{}
	resp, _ := client.Do(req)

	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Fprintf(w, string(body))

}

func main() {
	Router := readConfig()
	fmt.Printf("listening on port: %v\n", Router.Port)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		HttpHandler(w, r, Router)
	})
	http.ListenAndServe(":"+Router.Port, nil)
}

// readConfig reads configuration file written in json format, returns the Router struct
func readConfig() Router {
	file, e := ioutil.ReadFile("./config.json")
	if e != nil {
		fmt.Printf("File error: %v\n", e)
		os.Exit(1)
	}
	var root JsonRoot
	json.Unmarshal(file, &root)
	return root.Router
}

// convertRequest converts the incomming http.Request to the request specified in configuration file
func convertRequest(r *http.Request, forwardPath ForwardPath) *http.Request {
	req, _ := http.NewRequest(r.Method, forwardPath.Path, r.Body)
	req.Header.Set("Content-Type", forwardPath.ContentType)
	req.SetBasicAuth(forwardPath.BasicAuth.Username, forwardPath.BasicAuth.Password)

	return req
}

// converts http.Request with default values specified in configuration file
func convertRequestWithDefaultPath(r *http.Request, forwardPath DefaultForwardPath) *http.Request {
	req, _ := http.NewRequest(r.Method, forwardPath.Path, r.Body)
	req.Header.Set("Content-Type", forwardPath.ContentType)
	req.SetBasicAuth(forwardPath.BasicAuth.Username, forwardPath.BasicAuth.Password)
	return req
}
