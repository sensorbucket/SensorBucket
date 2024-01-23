package layout_utils

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type FlashMessage struct {
	Title       string
	Description string
	MessageType FlashMessageType
	CopyButton  bool
}

type FlashMessageType int

const (
	Warning FlashMessageType = iota
	Error
	Success
)

type FlashMessageRenderer func(msg FlashMessage) string

type FlashMessages []FlashMessage

type ctxKey int

const (
	ctxFlashMessagesKey ctxKey = iota
)

func (fm *FlashMessages) AsCookie() (http.Cookie, error) {
	b, err := json.Marshal(&fm)
	if err != nil {
		return http.Cookie{}, err
	}

	return http.Cookie{
		Name:     "flash_messages",
		Value:    base64.StdEncoding.WithPadding(base64.NoPadding).EncodeToString(b),
		Secure:   true,
		HttpOnly: true,
		Path:     "/",

		// If somehow the messages are not shown after 5 seconds they're not relevant anymore
		MaxAge: 5,
	}, nil
}

func ExtractFlashMessage(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if cookie, err := r.Cookie("flash_messages"); err == nil {

			// found flash messages to display to the user
			flashMessages := FlashMessages{}
			decoded, err := base64.StdEncoding.WithPadding(base64.NoPadding).DecodeString(cookie.Value)
			if err != nil {
				fmt.Printf("[Warning] found flash_messages cookie, but was not a valid base64 string\n")
			} else {
				err = json.Unmarshal(decoded, &flashMessages)
				if err != nil {
					fmt.Printf("[Warning] found flash_messages cookie, but couldnt convert to flash message: %s\n", err)
				} else {
					ctx = context.WithValue(ctx, ctxFlashMessagesKey, flashMessages)
				}
			}

			// unset the cookie, flash messages should only be shown once
			cookie.MaxAge = -1
			cookie.Path = "/"
			http.SetCookie(w, cookie)
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func FlashMessagesFromContext(ctx context.Context) (FlashMessages, bool) {
	if messages := ctx.Value(ctxFlashMessagesKey); messages != nil {
		if flashMessages, ok := messages.(FlashMessages); ok {
			return flashMessages, ok
		} else {
			log.Printf("[Warning] found flash messages in request context but couldnt parse\n")
		}
	}
	return FlashMessages{}, false
}

func AddSuccessFlashMessage(w http.ResponseWriter, r *http.Request, message string) context.Context {
	return addFlashMessage(w, r, "Success", message, Success, false)
}

func AddWarningFlashMessage(w http.ResponseWriter, r *http.Request, title string, message string, copyButton bool) context.Context {
	return addFlashMessage(w, r, title, message, Warning, copyButton)
}

func AddErrorFlashMessage(w http.ResponseWriter, r *http.Request, message string) context.Context {
	return addFlashMessage(w, r, "Error", message, Error, false)
}

func WriteSuccessFlashMessage(w http.ResponseWriter, r *http.Request, message string, renderer FlashMessageRenderer) {
	fm := getFlashMessage("Success", message, Success, false)
	w.Write([]byte(renderer(fm)))
}

func WriteWarningFlashMessage(w http.ResponseWriter, r *http.Request, title string, message string, copyButton bool, renderer FlashMessageRenderer) {
	fm := getFlashMessage("Warning", message, Warning, copyButton)
	w.Write([]byte(renderer(fm)))
}

func WriteErrorFlashMessage(w http.ResponseWriter, r *http.Request, message string, renderer FlashMessageRenderer) {
	fm := getFlashMessage("Error", message, Error, false)
	w.Write([]byte(renderer(fm)))
}

func addFlashMessage(w http.ResponseWriter, r *http.Request, title string, description string, msgType FlashMessageType, copyButton bool) context.Context {
	flashMessages := FlashMessages{}
	flashMessages, _ = FlashMessagesFromContext(r.Context())
	flashMessages = append(flashMessages, getFlashMessage(title, description, msgType, copyButton))
	newCtx := context.WithValue(r.Context(), ctxFlashMessagesKey, flashMessages)
	writeFlashMessagesToCookie(w, r.WithContext(newCtx))
	return newCtx
}

func getFlashMessage(title string, description string, msgType FlashMessageType, copyButton bool) FlashMessage {
	return FlashMessage{
		Title:       title,
		Description: description,
		MessageType: msgType,
		CopyButton:  copyButton,
	}
}

func writeFlashMessagesToCookie(w http.ResponseWriter, r *http.Request) {
	if messages, ok := FlashMessagesFromContext(r.Context()); ok {
		res, err := messages.AsCookie()
		if err != nil {
			log.Printf("[Warning] couldnt set flash_messages cookie %s\n", err)
		} else {
			http.SetCookie(w, &res)
		}
	} else {
		log.Printf("[Warning] no flash_messages could be found to be set as cookie\n")
	}
}
