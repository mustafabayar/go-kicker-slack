package main

import (
	"encoding/json"
	"fmt"
	"os"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/nlopes/slack"
)

var slackClient slack.Client

func main() {
	http.HandleFunc("/gokicker", slashCommandHandler)
	http.HandleFunc("/actions", actionHandler)
	http.ListenAndServe(os.Getenv("PORT"), nil)
}

func slashCommandHandler(w http.ResponseWriter, r *http.Request) {
	s, err := slack.SlashCommandParse(r)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	switch s.Command {
	case "/gokicker":
		attachments := []slack.Attachment{
			slack.Attachment{
				Title:      fmt.Sprintf("<@%s> joined!", s.UserID),
				Fallback:   "You are unable to answer this request",
				CallbackID: "kicker",
				Color:      "#20aa20",
				Actions: []slack.AttachmentAction{
					slack.AttachmentAction{
						Name:  "button-leave-0",
						Text:  "Leave",
						Type:  "button",
						Style: "danger",
						Value: s.UserID,
					},
				},
			},
			slack.Attachment{
				Title:      "Free Slot",
				Fallback:   "You are unable to answer this request",
				CallbackID: "kicker",
				Color:      "#20aa20",
				Actions: []slack.AttachmentAction{
					slack.AttachmentAction{
						Name:  "button-join-1",
						Text:  "Join",
						Type:  "button",
						Style: "primary",
						Value: "To join a match",
					},
				},
			},
			slack.Attachment{
				Title:      "Free Slot",
				Fallback:   "You are unable to answer this request",
				CallbackID: "kicker",
				Color:      "#20aa20",
				Actions: []slack.AttachmentAction{
					slack.AttachmentAction{
						Name:  "button-join-2",
						Text:  "Join",
						Type:  "button",
						Style: "primary",
						Value: "To join a match",
					},
				},
			},
			slack.Attachment{
				Title:      "Free Slot",
				Fallback:   "You are unable to answer this request",
				CallbackID: "kicker",
				Color:      "#20aa20",
				Actions: []slack.AttachmentAction{
					slack.AttachmentAction{
						Name:  "button-join-3",
						Text:  "Join",
						Type:  "button",
						Style: "primary",
						Value: "To join a match",
					},
				},
			},
		}
		message := slack.MsgOptionAttachments(attachments...)
		b, err := json.Marshal(message)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(b)

	default:
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func actionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Printf("[ERROR] Invalid method: %s", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("[ERROR] Failed to read request body: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	jsonStr, err := url.QueryUnescape(string(buf)[8:])
	if err != nil {
		log.Printf("[ERROR] Failed to unespace request body: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var message slack.AttachmentActionCallback
	err = json.Unmarshal([]byte(jsonStr), &message)
	if err != nil {
		log.Printf("[ERROR] Failed to decode json message from slack: %s", jsonStr)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	originalMessage := message.OriginalMessage

	var modifiedAttachment slack.Attachment
	var index int
	switch message.ActionCallback.AttachmentActions[0].Name {
	case "button-join-0", "button-leave-0":
		index = 0
		modifiedAttachment = originalMessage.Attachments[index]
		break
	case "button-join-1", "button-leave-1":
		index = 1
		modifiedAttachment = originalMessage.Attachments[index]
		break
	case "button-join-2", "button-leave-2":
		index = 2
		modifiedAttachment = originalMessage.Attachments[index]
		break
	case "button-join-3", "button-leave-3":
		index = 3
		modifiedAttachment = originalMessage.Attachments[index]
		break
	default:
		log.Printf("[ERROR] ]Invalid action was submitted: %s", message.ActionCallback.AttachmentActions[0].Name)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	action := modifiedAttachment.Actions[0]
	if strings.Contains(action.Name, "join") {
		modifiedAttachment.Color = "#20aa20" // Green
		modifiedAttachment.Actions = nil
		leaveAction := slack.AttachmentAction{
			Name:  fmt.Sprintf("button-leave-%v", index),
			Text:  "Leave",
			Type:  "button",
			Style: "danger",
			Value: message.User.ID,
		}
		modifiedAttachment.Actions = append(modifiedAttachment.Actions, leaveAction)
		modifiedAttachment.Title = fmt.Sprintf("<@%s> joined!", message.User.ID)
	}

	if strings.Contains(action.Name, "leave") {
		modifiedAttachment.Color = "#DCDCDC" // Gray
		modifiedAttachment.Actions = nil
		joinAction := slack.AttachmentAction{
			Name:  fmt.Sprintf("button-join-%v", index),
			Text:  "Join",
			Type:  "button",
			Style: "primary",
			Value: "To join a match",
		}
		modifiedAttachment.Actions = append(modifiedAttachment.Actions, joinAction)
		modifiedAttachment.Title = fmt.Sprintf("<@%s> left!", message.User.ID)
	}

	originalMessage.Attachments[index] = modifiedAttachment

	w.Header().Add("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&originalMessage)
}
