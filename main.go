package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/nlopes/slack"
)

var slackClient slack.Client

func main() {
	http.HandleFunc("/gokicker", slashCommandHandler)
	http.HandleFunc("/actions", actionHandler)
	http.ListenAndServe(":"+os.Getenv("PORT"), nil)
}

func slashCommandHandler(w http.ResponseWriter, r *http.Request) {
	s, err := slack.SlashCommandParse(r)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	switch s.Command {
	case "/gokicker":
		params := &slack.Msg{Text: s.Text}
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
		params.Attachments = attachments
		params.ResponseType = "in_channel"
		b, err := json.Marshal(params)
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
	var actionCallback slack.InteractionCallback
	log.Println(r.FormValue("payload"))
	err := json.Unmarshal([]byte(r.FormValue("payload")), &actionCallback)
	if err != nil {
		log.Println("[ERROR] Failed to decode json message from slack")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var modifiedAttachment slack.Attachment
	var index int
	actionName := actionCallback.ActionCallback.AttachmentActions[0].Name
	switch actionName {
	case "button-join-0", "button-leave-0":
		index = 0
		modifiedAttachment = actionCallback.OriginalMessage.Attachments[index]
		break
	case "button-join-1", "button-leave-1":
		index = 1
		modifiedAttachment = actionCallback.OriginalMessage.Attachments[index]
		break
	case "button-join-2", "button-leave-2":
		index = 2
		modifiedAttachment = actionCallback.OriginalMessage.Attachments[index]
		break
	case "button-join-3", "button-leave-3":
		index = 3
		modifiedAttachment = actionCallback.OriginalMessage.Attachments[index]
		break
	default:
		log.Printf("[ERROR] ]Invalid action was submitted: %s", actionName)
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
			Value: actionCallback.User.ID,
		}
		modifiedAttachment.Actions = append(modifiedAttachment.Actions, leaveAction)
		modifiedAttachment.Title = fmt.Sprintf("<@%s> joined!", actionCallback.User.ID)
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
		modifiedAttachment.Title = fmt.Sprintf("<@%s> left!", actionCallback.User.ID)
	}

	actionCallback.OriginalMessage.Attachments[index] = modifiedAttachment

	w.Header().Add("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&actionCallback.OriginalMessage)
}
