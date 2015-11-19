package main

import (
	"flag"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/tomyhero/billie/filter"
	"github.com/tomyhero/billie/notify"
	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"

	log "github.com/Sirupsen/logrus"
)

// configDir where the config file directory is.
var configDir string

func init() {
	// initialize values from flag
	flag.StringVar(&configDir, "config", "./assets/config/", "Path to the config dir ")
	flag.Parse()
}

func main() {
	// start server!
	goji.Post(regexp.MustCompile(`^/(?P<name>[a-zA-Z0-9_-]+)/(?P<form_name>[a-zA-Z0-9_-]+)/$`), handler)
	goji.Get("/__status/", status)
	goji.Serve()
}

func status(c web.C, w http.ResponseWriter, r *http.Request) {

	mem := &runtime.MemStats{}
	runtime.ReadMemStats(mem)
	memMb := float64((float64(mem.Alloc) / 1000) / 1000)
	unixtime := time.Now().Unix()

	line := []string{}
	line = append(line, fmt.Sprintf("%s %d %d", "num_goroutine", runtime.NumGoroutine(), unixtime))
	line = append(line, fmt.Sprintf("%s %f %d", "memory", memMb, unixtime))

	body := strings.Join(line, "\n")

	fmt.Fprintf(w, body)
}

func handler(c web.C, w http.ResponseWriter, r *http.Request) {

	config, err := getConfig(c)

	// fail to get config
	if err != nil {
		log.Error("fail to get config:", err)
		fmt.Fprintf(w, "SYSTEM ERROR")
		return
	}

	formConfig, err := getFormConfig(c, config)

	// fail to get form
	if err != nil {
		log.Error("fail to get form:", err)
		fmt.Fprintf(w, "SYSTEM ERROR")
		return
	}

	// getting fields and attachments data.

	allowFields, allowFileExtentions := getAllowSettings(formConfig)
	fields, attachments, err := getData(r, allowFields, allowFileExtentions)

	// failt to getting data
	if err != nil {
		log.Error(err)
		http.Redirect(w, r, formConfig["error"].(string), http.StatusFound)
		return
	}

	// get notify setting
	notifyList, hasNotifies := formConfig["notifies"].(string)

	if hasNotifies {
		for _, part := range strings.Split(notifyList, ",") {
			p := strings.Split(part, ".")
			notifyType := p[0]
			notifyName := p[1]

			// get notify setting
			setting, hasSetting := config["notify"].(map[string]interface{})[notifyType].(map[string]interface{})[notifyName].(map[string]interface{})

			if !hasSetting {
				log.Error("Can not found notify Data:", part)
				fmt.Fprintf(w, "SYSTEM ERROR")
				return
			}

			// get filter setting
			filterConfig, hasFilterConfig := config["filter"].(map[string]interface{})[setting["filter"].(string)].(map[string]interface{})

			if !hasFilterConfig {
				log.Error("Can not find filter:", setting["filter"].(string))
				fmt.Fprintf(w, "SYSTEM ERROR")
				return
			}

			filterFormat, hasFormat := filterConfig["format"].(string)

			if !hasFormat {
				log.Error("format is empty", setting["filter"].(string))
				fmt.Fprintf(w, "SYSTEM ERROR")
				return
			}

			//  ok to get data!
			f := getFilterFormat(filterFormat, config)
			body := f.Parse(fields, attachments)

			// notify!
			n := createNotifyObject(notifyType, filterFormat, formConfig["title"].(string), setting)
			n.Notify(body, attachments)
		}
	}

	// everything fine!
	http.Redirect(w, r, formConfig["success"].(string), http.StatusFound)
}

func getConfig(c web.C) (map[string]interface{}, error) {
	name := c.URLParams["name"]
	var config map[string]interface{}
	if _, err := toml.DecodeFile(configDir+name+".toml", &config); err != nil {
		return nil, err
	}
	return config, nil
}

func getFormConfig(c web.C, config map[string]interface{}) (map[string]interface{}, error) {
	formName := c.URLParams["form_name"]

	formConfigs, hasFormConfigs := config["receiver"].(map[string]interface{})["web"].(map[string]interface{})["form"].(map[string]interface{})

	if !hasFormConfigs {
		return nil, fmt.Errorf("Can not found form config")
	}

	formConfig, hasFormConfig := formConfigs[formName].(map[string]interface{})

	if !hasFormConfig {
		return nil, fmt.Errorf("can not found form")
	}

	return formConfig, nil
}

func getAllowSettings(formConfig map[string]interface{}) (map[string]bool, map[string]bool) {

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

	return allowFields, allowFileExtentions
}

func createNotifyObject(notifyType string, filterFormat string, title string, setting map[string]interface{}) notify.NotifyExecutor {

	var n notify.NotifyExecutor

	if notifyType == "email" {

		n = &notify.Email{
			From:        setting["from"].(string),
			To:          setting["to"].(string),
			CC:          setting["cc"].(string),
			Title:       title,
			ContentType: getContentType(filterFormat),
			SMTP:        setting["smtp"].(map[string]interface{}),
		}
	} else {

		n = &notify.Slack{
			Token:       setting["token"].(string),
			Channel:     setting["channel"].(string),
			Username:    setting["username"].(string),
			AsUser:      setting["as_user"].(bool),
			UnfurlLinks: setting["unfurl_links"].(bool),
			UnfurlMedia: setting["unfurl_media"].(bool),
			IconURL:     setting["icon_url"].(string),
			IconEmoji:   setting["icon_emoji"].(string),
		}
	}
	return n
}

func getData(r *http.Request, allowFields map[string]bool, allowFileExtentions map[string]bool) (map[string]interface{}, map[string][]*multipart.FileHeader, error) {

	fields := map[string]interface{}{}
	attachments := map[string][]*multipart.FileHeader{}

	err := r.ParseMultipartForm(1024 * 1024)

	onMultipartForm := true

	if err == http.ErrNotMultipart {
		onMultipartForm = false
	} else if err != nil {
		return nil, nil, err
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

	return fields, attachments, nil
}

func getFilterName(config map[string]interface{}) (filterName string) {
	filterName = "text"

	filterConfig, hasFilter := config["filter"].(map[string]interface{})
	if hasFilter {
		format, hasFormat := filterConfig["format"].(string)
		if hasFormat {
			filterName = format
		}
	}

	return filterName
}

func getFilterFormat(filterName string, config map[string]interface{}) (f filter.FilterExecutor) {
	f = &filter.Text{}

	switch filterName {
	case "html":
		htmlFilter := filter.HTML{}
		filterConfig, hasFilter := config["filter"].(map[string]interface{})
		if hasFilter {
			htmlConfig, hasHTML := filterConfig["html"].(map[string]interface{})
			if hasHTML {
				template, hasTemplate := htmlConfig["template"].(string)
				if hasTemplate {
					htmlFilter.Template = template
				}
			}
		}
		f = &htmlFilter
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
