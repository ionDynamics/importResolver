package main

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/fcgi"
	"runtime"
	"strings"

	"github.com/codegangsta/negroni"
	idl "go.iondynamics.net/iDlogger"
	"go.iondynamics.net/iDnegroniLog"
)

type config struct {
	Fcgi   bool
	Listen string
	Match  map[string]string
}

var Config config
var Tpl *template.Template

func init() {
	byt, err := ioutil.ReadFile("config.json")

	if err == nil {
		err = json.Unmarshal(byt, &Config)
	}

	if err == nil {
		Tpl, err = template.ParseFiles("template.tpl")
	}

	if err != nil {
		idl.Emerg("init", err)
	}

}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	logger := iDnegroniLog.NewMiddleware(idl.StandardLogger())

	n := negroni.New(logger, negroni.NewStatic(http.Dir("public")))
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleFunc)
	n.UseHandler(mux)

	if Config.Fcgi {
		listener, err := net.Listen("tcp", Config.Listen)
		if err != nil {
			idl.Emerg(err)
		}
		fcgi.Serve(listener, n)
	} else {
		n.Run(Config.Listen)
	}
}

func handleFunc(w http.ResponseWriter, req *http.Request) {
	requested := req.Host + req.RequestURI
	for canonical, destination := range Config.Match {
		if strings.HasPrefix(requested, canonical) {
			path := strings.TrimPrefix(requested, canonical)
			Tpl.Execute(w, map[string]string{
				"Canonical":   requested,
				"Destination": destination + path,
			})
		}
	}
}
