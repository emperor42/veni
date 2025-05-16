package veni

import (
	"fmt"
	"net/http"
	"path"
	"strings"
)

type VeniHandler interface {
	Process(w http.ResponseWriter, r *http.Request)
	AddRoute(routeName string, handle func(http.ResponseWriter, *http.Request))
}

type VeniContext struct {
	Name       string
	ConnectAPI VeniHandler
	PatchAPI   VeniHandler
	PostAPI    VeniHandler
	DeleteAPI  VeniHandler
	PutAPI     VeniHandler
	TraceAPI   VeniHandler
	OptionsAPI VeniHandler
	HeadAPI    VeniHandler
	GetAPI     VeniHandler
}

func (v *VeniContext) Process(w http.ResponseWriter, r *http.Request) {
	base := path.Base(r.URL.Path)
	switch strings.ToLower(base) {
	case "get":
		v.GetAPI.Process(w, r)
	case "head":
		v.HeadAPI.Process(w, r)
	case "options":
		v.OptionsAPI.Process(w, r)
	case "trace":
		v.TraceAPI.Process(w, r)
	case "put":
		v.PutAPI.Process(w, r)
	case "delete":
		v.DeleteAPI.Process(w, r)
	case "post":
		v.PostAPI.Process(w, r)
	case "patch":
		v.PatchAPI.Process(w, r)
	case "connect":
		v.ConnectAPI.Process(w, r)
	}
}

func (v *VeniContext) Comply(r *http.Request) bool {
	base := path.Base(r.URL.Path)
	switch strings.ToLower(base) {
	case "get", "head", "options", "trace", "put", "delete", "post", "patch", "connect":
		return true
	default:
		return false
	}
}

func (v *VeniContext) ProcessHeader() {
	fmt.Println("temp")
}

func (v *VeniContext) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.RequestURI() != "/call" {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}

	if r.Method != "GET" {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, v.Name)
	}

	fmt.Fprintf(w, "Call Complete!")
}
