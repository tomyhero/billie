package main

import (
	"fmt"
	"github.com/BurntSushi/toml"
	log "github.com/Sirupsen/logrus"
	"github.com/tomyhero/billie/filter"
	"github.com/tomyhero/billie/notify"
	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"
	"mime/multipart"
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
	if _, err := toml.DecodeFile("./assets/config/"+name+".toml", &config); err != nil {
		log.Error(err)
		fmt.Fprintf(w, "SYSTEM ERROR")
		return
	}

	err := r.ParseMultipartForm(1024 * 10)

	onMultipartForm := true

	fields := map[string]interface{}{}
	attachments := map[string][]*multipart.FileHeader{}
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
			log.Info(name, v)
			fields[name] = v
		}
	} else {

		for name, v := range r.PostForm {
			fields[name] = v
		}
	}

	f := &filter.Text{}
	body := f.Parse(fields, attachments)

	notifyConfig, hasNotify := config["notify"].(map[string]interface{})
	if hasNotify {
		emailConfig, hasEmail := notifyConfig["email"].([]map[string]interface{})
		if hasEmail {
			for _, setting := range emailConfig {

				n := notify.Email{
					From:  setting["from"].(string),
					To:    setting["to"].(string),
					Title: setting["title"].(string),
					SMTP:  setting["smtp"].(map[string]interface{}),
				}
				n.Notify(body, attachments)
			}
		}
	}

	fmt.Fprintf(w, body)
}
