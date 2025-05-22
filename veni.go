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
	RemoveRoute(routeName string)
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

func (v *VeniContext) AddRoute(routeName string, handle func(http.ResponseWriter, *http.Request)) {
	v.ConnectAPI.AddRoute(routeName, handle)
	v.PatchAPI.AddRoute(routeName, handle)
	v.PostAPI.AddRoute(routeName, handle)
	v.DeleteAPI.AddRoute(routeName, handle)
	v.PutAPI.AddRoute(routeName, handle)
	v.TraceAPI.AddRoute(routeName, handle)
	v.OptionsAPI.AddRoute(routeName, handle)
	v.HeadAPI.AddRoute(routeName, handle)
	v.GetAPI.AddRoute(routeName, handle)
}

func (v *VeniContext) RemoveRoute(routeName string) {
	v.ConnectAPI.RemoveRoute(routeName)
	v.PatchAPI.RemoveRoute(routeName)
	v.PostAPI.RemoveRoute(routeName)
	v.DeleteAPI.RemoveRoute(routeName)
	v.PutAPI.RemoveRoute(routeName)
	v.TraceAPI.RemoveRoute(routeName)
	v.OptionsAPI.RemoveRoute(routeName)
	v.HeadAPI.RemoveRoute(routeName)
	v.GetAPI.RemoveRoute(routeName)
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
