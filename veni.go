package veni

import (
	"fmt"
	"net/http"
	"path"
	"strings"
)

type VeniHandler interface {
	Process(w http.ResponseWriter, r *http.Request) http.ResponseWriter
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

func (v *VeniContext) Process(w http.ResponseWriter, r *http.Request) http.ResponseWriter {
	base := path.Base(r.URL.Path)
	switch strings.ToLower(base) {
	case "get":
		return v.GetAPI.Process(w, r)
	case "head":
		return v.HeadAPI.Process(w, r)
	case "options":
		return v.OptionsAPI.Process(w, r)
	case "trace":
		return v.TraceAPI.Process(w, r)
	case "put":
		return v.PutAPI.Process(w, r)
	case "delete":
		return v.DeleteAPI.Process(w, r)
	case "post":
		return v.PostAPI.Process(w, r)
	case "patch":
		return v.PatchAPI.Process(w, r)
	case "connect":
		return v.ConnectAPI.Process(w, r)
	default:
		return w
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
