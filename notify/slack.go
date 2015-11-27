package notify

import (
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"os"

	"github.com/nlopes/slack"

	log "github.com/Sirupsen/logrus"
)

// Slack is setting for posting on slack.
type Slack struct {
	Token       string
	Channel     string
	Username    string
	AsUser      bool
	UnfurlLinks bool
	UnfurlMedia bool
	IconURL     string
	IconEmoji   string
}

// Notify is function to notify slack.
func (s *Slack) Notify(body string, attachments [][]*multipart.FileHeader) {
	// auth
	api := slack.New(s.Token)

	for _, tmp := range attachments {
		for _, attachment := range tmp {
			fileURL, err := upSlack(api, attachment)
			if err != nil {
				log.Errorf("upload slack error: %v", err)
			} else {
				// add url to body
				body += fmt.Sprintln(fileURL)
			}
		}
	}

	api.PostMessage(s.Channel, body, slack.PostMessageParameters{
		Username:    s.Username,
		AsUser:      s.AsUser,
		UnfurlLinks: s.UnfurlLinks,
		UnfurlMedia: s.UnfurlMedia,
		IconURL:     s.IconURL,
		IconEmoji:   s.IconEmoji,
	})

}

func upSlack(api *slack.Client, attachment *multipart.FileHeader) (string, error) {
	upFile, err := ioutil.TempFile("", "upSlack_")
	defer os.Remove(upFile.Name())

	f, err := attachment.Open()
	if err != nil {
		return "", fmt.Errorf("attached file open error: %v", err)
	}

	// file save
	written, err := io.Copy(upFile, f)
	if err != nil {
		return "", fmt.Errorf("file save error: %v, written: %d", err, written)
	}

	fileInfo, err := api.UploadFile(slack.FileUploadParameters{
		File:     upFile.Name(),
		Filename: attachment.Filename,
	})
	if err != nil {
		return "", fmt.Errorf("file upload error: %v", err)
	}

	return fileInfo.URL, nil
}
