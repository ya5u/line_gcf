package linegcf

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/firestore"
)

var (
	projectID string = os.Getenv("GCP_PROJECT")
	ctx       context.Context
	client    *firestore.Client
)

func init() {
	log.Printf("init - projectID: %s", projectID)

	var err error

	ctx = context.Background()
	client, err = firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("firestore.NewClient: %v", err)
	}
}

var (
	lineChannelSecret string = os.Getenv("LINE_CHANNEL_SECRET")
	req               struct {
		Destination string `json:"destination"`
		Events      []Event
	}
)

// Event Object
type Event struct {
	Type       string `json:"type"`
	Timestamp  int64  `json:"timestamp"`
	ReplyToken string `json:"replyToken"`
	Source     LineSource
	Message    LineMessage
}

// LineSource Object
type LineSource struct {
	Type    string `json:"type"`
	UserID  string `json:"userId"`
	GroupID string `json:"groupId"`
	RoomID  string `json:"roomId"`
}

// LineMessage Object
type LineMessage struct {
	ID              string  `json:"id"`
	Type            string  `json:"type"`
	Text            string  `json:"text"`
	PackageID       string  `json:"packageId"`
	StickerID       string  `json:"stickerId"`
	FileName        string  `json:"fileName"`
	FileSize        int64   `json:"fileSize"`
	Title           string  `json:"title"`
	Address         string  `json:"address"`
	Latitude        float64 `json:"latitude"`
	Longitude       float64 `json:"longitude"`
	Duration        int64   `json:"duration"`
	ContentProvider struct {
		Type               string `json:"type"`
		OriginalContentURL string `json:"originalContentUrl"`
		PreviewImageURL    string `json:"previewImageUrl"`
	}
}

// Handler for Cloud Functions
func Handler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Content-Type: %s", r.Header.Get("Content-Type"))

	defer r.Body.Close()
	var bodyBytes []byte
	bodyBytes, _ = ioutil.ReadAll(r.Body)

	verified := verifySignature(r.Header.Get("X-Line-Signature"), bodyBytes)
	if !verified {
		log.Printf("Signature is NOT Verified")
		w.WriteHeader(http.StatusForbidden)
		return
	}

	r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("JSON parse ERROR: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	d := req.Destination
	events := req.Events

	log.Printf("destination: %s", d)

	for _, e := range events {
		log.Printf("Timestamp: %d", e.Timestamp)
		log.Printf("UserID: %s", e.Source.UserID)

		switch e.Message.Type {
		case "text":
			log.Printf("Text: %s", e.Message.Text)
		case "video":
			fallthrough
		case "audio":
			log.Printf("Duration: %d", e.Message.Duration)
			fallthrough
		case "image":
			log.Printf("ContetnProvider.Type: %s", e.Message.ContentProvider.Type)
		case "file":
			log.Printf("FileName: %s", e.Message.FileName)
			log.Printf("FileSize: %d", e.Message.FileSize)
		case "location":
			log.Printf("Title: %s", e.Message.Title)
			log.Printf("Address: %s", e.Message.Address)
			log.Printf("Latitude: %f", e.Message.Latitude)
			log.Printf("Longitude: %f", e.Message.Longitude)
		case "sticker":
			log.Printf("PackageID: %s", e.Message.PackageID)
			log.Printf("StickerID: %s", e.Message.StickerID)
		default:
			log.Printf("Type: %s", e.Message.Type)
		}

		go writeMsg(e.Timestamp, e.ReplyToken, e.Source, e.Message)
	}

	fmt.Fprintf(w, "Ok")
}

func verifySignature(signature string, body []byte) bool {
	decoded, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		log.Printf("Base64 decode ERROR: %v", err)
		return false
	}

	hash := hmac.New(sha256.New, []byte(lineChannelSecret))
	hash.Write(body)
	return hmac.Equal(decoded, hash.Sum(nil))
}

// FsMessage Document in Firestore collection "Messages"
type FsMessage struct {
	Timestamp          int64   `firestore:"timestamp"`
	ReplyToken         string  `firestore:"replyToken"`
	SourceType         string  `firestore:"sourceType"`
	UserID             string  `firestore:"userID"`
	GroupID            string  `firestore:"groupID"`
	RoomID             string  `firestore:"roomID"`
	Type               string  `firestore:"type"`
	Text               string  `firestore:"text"`
	PackageID          string  `firestore:"packageID"`
	StickerID          string  `firestore:"stickerID"`
	FileName           string  `firestore:"fileName"`
	FileSize           int64   `firestore:"fileSize"`
	Title              string  `firestore:"title"`
	Address            string  `firestore:"address"`
	Latitude           float64 `firestore:"latitude"`
	Longitude          float64 `firestore:"longitude"`
	Duration           int64   `firestore:"duration"`
	ContentType        string  `firestore:"contentType"`
	OriginalContentURL string  `firestore:"originalContentURL"`
	PreviewImageURL    string  `firestore:"previewImageURL"`
}

func writeMsg(timestamp int64, replyToken string, source LineSource, msg LineMessage) {
	log.Printf("writeMsg - timestamp:%d source:%v msg:%v", timestamp, source, msg)

	doc := client.Doc("Messages/" + msg.ID)
	wr, err := doc.Create(ctx, FsMessage{
		Timestamp:          timestamp,
		ReplyToken:         replyToken,
		SourceType:         source.Type,
		UserID:             source.UserID,
		GroupID:            source.GroupID,
		RoomID:             source.RoomID,
		Type:               msg.Type,
		Text:               msg.Text,
		PackageID:          msg.PackageID,
		StickerID:          msg.StickerID,
		FileName:           msg.FileName,
		FileSize:           msg.FileSize,
		Title:              msg.Title,
		Address:            msg.Address,
		Latitude:           msg.Latitude,
		Longitude:          msg.Longitude,
		Duration:           msg.Duration,
		ContentType:        msg.ContentProvider.Type,
		OriginalContentURL: msg.ContentProvider.OriginalContentURL,
		PreviewImageURL:    msg.ContentProvider.PreviewImageURL,
	})
	if err != nil {
		log.Fatalf("writeMsg: %v", err)
	}
	log.Printf("WriteResult: %v", wr)
}
