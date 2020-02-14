package main

import (
	"bytes"
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
	http.HandleFunc("/slash", slashCommandHandler)
	http.HandleFunc("/interactive", actionHandler)
	http.ListenAndServe(":"+os.Getenv("PORT"), nil)
}

func slashCommandHandler(w http.ResponseWriter, r *http.Request) {
	s, err := slack.SlashCommandParse(r)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	switch s.Command {
	case "/kicker":
		params := &slack.Msg{}
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
				Color:      "#DCDCDC",
				Actions: []slack.AttachmentAction{
					slack.AttachmentAction{
						Name:  "button-join-1",
						Text:  "Join",
						Type:  "button",
						Style: "primary",
						Value: "",
					},
				},
			},
			slack.Attachment{
				Title:      "Free Slot",
				Fallback:   "You are unable to answer this request",
				CallbackID: "kicker",
				Color:      "#DCDCDC",
				Actions: []slack.AttachmentAction{
					slack.AttachmentAction{
						Name:  "button-join-2",
						Text:  "Join",
						Type:  "button",
						Style: "primary",
						Value: "",
					},
				},
			},
			slack.Attachment{
				Title:      "Free Slot",
				Fallback:   "You are unable to answer this request",
				CallbackID: "kicker",
				Color:      "#DCDCDC",
				Actions: []slack.AttachmentAction{
					slack.AttachmentAction{
						Name:  "button-join-3",
						Text:  "Join",
						Type:  "button",
						Style: "primary",
						Value: "",
					},
				},
			},
		}

		params.Text = "New kicker game created. Feel free to join!"
		params.Attachments = attachments
		params.ResponseType = slack.ResponseTypeInChannel
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
	err := json.Unmarshal([]byte(r.FormValue("payload")), &actionCallback)
	if err != nil {
		log.Println("[ERROR] Failed to decode json message from slack")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var selectedAttachment slack.Attachment
	var index int
	var originalMessage = actionCallback.OriginalMessage
	actionName := actionCallback.ActionCallback.AttachmentActions[index].Name
	switch actionName {
	case "button-join-0", "button-leave-0":
		selectedAttachment = originalMessage.Attachments[index]
		break
	case "button-join-1", "button-leave-1":
		index = 1
		selectedAttachment = originalMessage.Attachments[index]
		break
	case "button-join-2", "button-leave-2":
		index = 2
		selectedAttachment = originalMessage.Attachments[index]
		break
	case "button-join-3", "button-leave-3":
		index = 3
		selectedAttachment = originalMessage.Attachments[index]
		break
	default:
		log.Printf("[ERROR] Invalid action was submitted: %s", actionName)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var playerIds []string
	for _, a := range originalMessage.Attachments {
		if a.Actions[0].Value != "" {
			playerIds = append(playerIds, a.Actions[0].Value)
		}
	}

	action := selectedAttachment.Actions[0]
	if strings.Contains(action.Name, "join") {
		samePerson := false
		for _, p := range playerIds {
			if p == actionCallback.User.ID {
				warnUser := &slack.Msg{}
				warnUser.Text = "You are already occupying a spot in this game, leave some room to others!"
				warnUser.ReplaceOriginal = false
				warnUser.DeleteOriginal = false
				warnUser.ResponseType = slack.ResponseTypeEphemeral
				w.Header().Add("Content-type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(&warnUser)
				samePerson = true
			}
		}

		if samePerson {
			return
		}

		selectedAttachment.Color = "#20aa20" // Green
		selectedAttachment.Actions = nil
		leaveAction := slack.AttachmentAction{
			Name:  fmt.Sprintf("button-leave-%v", index),
			Text:  "Leave",
			Type:  "button",
			Style: "danger",
			Value: actionCallback.User.ID,
		}
		selectedAttachment.Actions = []slack.AttachmentAction{leaveAction}
		selectedAttachment.Title = fmt.Sprintf("<@%s> joined!", actionCallback.User.ID)
		playerIds = append(playerIds, actionCallback.User.ID)
	}

	if strings.Contains(action.Name, "leave") {
		if selectedAttachment.Actions[0].Value != actionCallback.User.ID {
			warnUser := &slack.Msg{}
			warnUser.Text = "You can not leave a spot that belongs to someone else!"
			warnUser.ReplaceOriginal = false
			warnUser.DeleteOriginal = false
			warnUser.ResponseType = slack.ResponseTypeEphemeral
			w.Header().Add("Content-type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(&warnUser)
			return
		}

		selectedAttachment.Color = "#DCDCDC" // Gray
		selectedAttachment.Actions = nil
		joinAction := slack.AttachmentAction{
			Name:  fmt.Sprintf("button-join-%v", index),
			Text:  "Join",
			Type:  "button",
			Style: "primary",
			Value: "",
		}
		selectedAttachment.Actions = []slack.AttachmentAction{joinAction}
		selectedAttachment.Title = fmt.Sprintf("<@%s> left!", actionCallback.User.ID)
		playerIds = playerIds[:len(playerIds)-1]
	}

	originalMessage.Attachments[index] = selectedAttachment
	originalMessage.ResponseType = slack.ResponseTypeInChannel
	originalMessage.ReplaceOriginal = true

	switch len(playerIds) {
	case 4:
		originalMessage.Text = "Enjoy the game! :soccer:"
		originalMessage.Attachments = nil

		reply := &slack.Msg{}
		reply.Text = fmt.Sprintf("<@%s>, <@%s>, <@%s>, <@%s> GO GO GO!", playerIds[0], playerIds[1], playerIds[2], playerIds[3])
		reply.ReplaceOriginal = false
		reply.ThreadTimestamp = actionCallback.MessageTs
		reply.ResponseType = slack.ResponseTypeInChannel

		w.Header().Add("Content-type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(&reply)

		jsonValue, err := json.Marshal(originalMessage)
		if err != nil {
			log.Println("[ERROR] Failed to encode json message for Slack")
		}

		resp, err := http.Post(actionCallback.ResponseURL, "application/json", bytes.NewBuffer(jsonValue))
		if err != nil || resp.StatusCode != 200 {
			log.Println("[ERROR] Slack request was unsuccessful:", resp.StatusCode)
		}
		return
	case 0:
		originalMessage.Text = "This game has been cancelled!"
		originalMessage.Attachments = nil
	}

	w.Header().Add("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&originalMessage)
}
