package main

import (
	"flag"
	"fmt"
	"mime/multipart"
	"net/http"
	"regexp"

	"github.com/BurntSushi/toml"
	"github.com/tomyhero/billie/filter"
	"github.com/tomyhero/billie/notify"
	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"

	log "github.com/Sirupsen/logrus"
)

var configDir string

func main() {

	flag.StringVar(&configDir, "config", "./assets/config/", "Path to the config dir ")
	flag.Parse()

	goji.Post(regexp.MustCompile(`^/(?P<name>[a-zA-Z0-9_-]+)/$`), handler)
	goji.Serve()
}

func handler(c web.C, w http.ResponseWriter, r *http.Request) {

	name := c.URLParams["name"]

	var config map[string]interface{}
	if _, err := toml.DecodeFile(configDir+name+".toml", &config); err != nil {
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

		// slack
		slackConfig, hasSlack := notifyConfig["slack"].([]map[string]interface{})
		if hasSlack {
			for _, setting := range slackConfig {

				n := notify.Slack{
					Token:       setting["token"].(string),
					Channel:     setting["channel"].(string),
					Username:    setting["username"].(string),
					AsUser:      setting["as_user"].(bool),
					UnfurlLinks: setting["unfurl_links"].(bool),
					UnfurlMedia: setting["unfurl_media"].(bool),
					IconURL:     setting["icon_url"].(string),
					IconEmoji:   setting["icon_emoji"].(string),
				}
				n.Notify(body, attachments)
			}
		}
	}

	fmt.Fprintf(w, body)
}
