package main

import (
	"fmt"
	"github.com/BurntSushi/toml"
	log "github.com/Sirupsen/logrus"
	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"
	"net/http"
	"regexp"
)

func main() {
	goji.Post(regexp.MustCompile(`^/(?P<name>[a-zA-Z0-9_-]+)/$`), handler)
	goji.Serve()
}

func handler(c web.C, w http.ResponseWriter, r *http.Request) {

	name := c.URLParams["name"]

	var config map[string]interface{}
	if _, err := toml.DecodeFile("./assets/"+name+".toml", &config); err != nil {
		log.Error(err)
		fmt.Fprintf(w, "SYSTEM ERROR")
		return
	}

	err := r.ParseMultipartForm(1024 * 10)

	onMultipartForm := true

	fields := map[string]interface{}{}
	attachments := map[string]interface{}{}
	if err == http.ErrNotMultipart {
		onMultipartForm = false
	} else if err != nil {
		log.Error(err)
		fmt.Fprintf(w, "SYSTEM ERROR")
		return
	}

	if onMultipartForm {
		for name, f := range r.MultipartForm.File {
			attachments[name] = f
		}

		for name, v := range r.MultipartForm.Value {
			fields[name] = v
		}
	} else {

		for name, v := range r.PostForm {
			fields[name] = v
		}
	}

	log.Info("field", fields)
	log.Info("attachments", attachments)
	log.Info(config)

	fmt.Fprintf(w, name+" Hello, World")
}
