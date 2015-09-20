package main

import (
	"flag"
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/tomyhero/billie/filter"
	"github.com/tomyhero/billie/notify"
	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"

	log "github.com/Sirupsen/logrus"
)

var configDir string

type filterFormat interface {
	Parse(map[string]interface{}, map[string][]*multipart.FileHeader) string
}

func main() {

	flag.StringVar(&configDir, "config", "./assets/config/", "Path to the config dir ")
	flag.Parse()

	goji.Post(regexp.MustCompile(`^/(?P<name>[a-zA-Z0-9_-]+)/(?P<form_name>[a-zA-Z0-9_-]+)/$`), handler)
	goji.Serve()
}

func handler(c web.C, w http.ResponseWriter, r *http.Request) {

	name := c.URLParams["name"]
	formName := c.URLParams["form_name"]

	var config map[string]interface{}
	if _, err := toml.DecodeFile(configDir+name+".toml", &config); err != nil {
		log.Error(err)
		fmt.Fprintf(w, "SYSTEM ERROR")
		return
	}

	formConfigs := config["receiver"].(map[string]interface{})["web"].(map[string]interface{})["form"].([]map[string]interface{})

	var formConfig map[string]interface{}

	for _, setting := range formConfigs {
		if formName == setting["name"].(string) {
			formConfig = map[string]interface{}{}
			formConfig = setting
		}
	}

	if formConfig == nil {
		fmt.Fprintf(w, "FORM NOT FOUND")
		return
	}

	supportedFields, hasSupportedFields := formConfig["supported_fields"].(string)

	allowFields := map[string]bool{}
	if hasSupportedFields {
		check := strings.Split(supportedFields, ",")
		for _, v := range check {
			allowFields[v] = true
		}
	}

	supportedFileExtentions, hasSupportedFileExtentions := formConfig["supported_file_extentions"].(string)

	allowFileExtentions := map[string]bool{}
	if hasSupportedFileExtentions {
		check := strings.Split(supportedFileExtentions, ",")
		for _, v := range check {
			allowFileExtentions[v] = true
		}
	}

	err := r.ParseMultipartForm(1024 * 1024)

	onMultipartForm := true

	fields := map[string]interface{}{}
	attachments := map[string][]*multipart.FileHeader{}
	if err == http.ErrNotMultipart {
		onMultipartForm = false
	} else if err != nil {
		log.Error(err)
		http.Redirect(w, r, formConfig["error"].(string), http.StatusFound)
		return
	}

	if onMultipartForm {
		for name, f := range r.MultipartForm.File {
			tmp := []*multipart.FileHeader{}
			for _, a := range f {
				_, allowdExt := allowFileExtentions[filepath.Ext(a.Filename)]
				if allowdExt {
					tmp = append(tmp, a)
				}
			}
			_, allowd := allowFields[name]
			if allowd && len(tmp) > 0 {
				attachments[name] = tmp
			}
		}

		for name, v := range r.MultipartForm.Value {
			_, allowd := allowFields[name]
			if allowd {
				fields[name] = v
			}
		}
	} else {

		for name, v := range r.PostForm {
			_, allowd := allowFields[name]
			if allowd {
				fields[name] = v
			}
		}
	}

	filterName := getFilterName(config)
	f := getFilterFormat(filterName)

	body := f.Parse(fields, attachments)

	notifyConfig, hasNotify := config["notify"].(map[string]interface{})
	if hasNotify {
		emailConfig, hasEmail := notifyConfig["email"].([]map[string]interface{})
		if hasEmail {
			for _, setting := range emailConfig {

				n := notify.Email{
					From:        setting["from"].(string),
					To:          setting["to"].(string),
					Title:       setting["title"].(string),
					ContentType: getContentType(filterName),
					SMTP:        setting["smtp"].(map[string]interface{}),
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

	http.Redirect(w, r, formConfig["success"].(string), http.StatusFound)
}

func getFilterName(config map[string]interface{}) (filterName string) {
	filterName = "text"

	filterConfig, hasFilter := config["filter"].(map[string]interface{})
	if hasFilter {
		formatConfig, hasFormat := filterConfig["format"].(string)
		if hasFormat {
			filterName = formatConfig
		}
	}

	return filterName
}

func getFilterFormat(filterName string) (f filterFormat) {
	f = &filter.Text{}

	switch filterName {
	case "html":
		f = &filter.HTML{}
	}

	return f
}

func getContentType(filterName string) (contentType string) {
	contentType = "text/plain"

	switch filterName {
	case "html":
		contentType = "text/html"
	}

	log.Info(contentType)
	return contentType
}
