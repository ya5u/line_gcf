package linegcf

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/firestore"
)

var projectID string = os.Getenv("GCP_PROJECT")

var (
	ctx context.Context
	client *firestore.Client
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

var req struct {
	Destination string `json: "destination"`
	Events      []Event
}

// Event Object
type Event struct {
	Type       string `json: "type"`
	Timestamp  int64  `json: "timestamp"`
	ReplyToken string `json: "replyToken"`
	Source     struct {
		Type    string `json: "type"`
		UserID  string `json: "userId"`
		GroupID string `json: "groupId"`
		RoomID  string `json: roomId`
	}
	Message LineMessage
}

// LineMessage Object
type LineMessage struct {
	ID              string  `json: "id"`
	Type            string  `json: "type"`
	Text            string  `json: "text"`
	PackageID       string  `json: "packageId"`
	StickerID       string  `json: "stickerId"`
	FileName        string  `json: "fileName"`
	FileSize        int64   `json: "fileSize"`
	Title           string  `json: "title"`
	Address         string  `json: "address"`
	Latitude        float64 `json: "latitude"`
	Longitude       float64 `json: "longitude"`
	Duration        int64   `json: "duration"`
	ContentProvider struct {
		Type               string `json: "type"`
		OriginalContentURL string `json: "originalContentUrl"`
		PreviewImageURL    string `json: "previewImageUrl"`
	}
}

// Handler for Cloud Functions
func Handler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Content-Type: %s", r.Header.Get("Content-Type"))

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("JSON parse ERROR: %v", err)
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

		go writeMsg(e.Timestamp, e.Message)
	}

	fmt.Fprintf(w, "Ok")
}

// FsMessage Document in Firestore collection "Messages"
type FsMessage struct {
	Timestamp          int64   `firestore: "timestamp"`
	Type               string  `firestore: "type"`
	Text               string  `firestore: "text"`
	PackageID          string  `firestore: "packageId"`
	StickerID          string  `firestore: "stickerId"`
	FileName           string  `firestore: "fileName"`
	FileSize           int64   `firestore: "fileSize"`
	Title              string  `firestore: "title"`
	Address            string  `firestore: "address"`
	Latitude           float64 `firestore: "latitude"`
	Longitude          float64 `firestore: "longitude"`
	Duration           int64   `firestore: "duration"`
	ContentType        string  `firestore: "contentType"`
	OriginalContentURL string  `firestore: "originalContentUrl"`
	PreviewImageURL    string  `firestore: "previewImageUrl"`
}

func writeMsg(timestamp int64, msg LineMessage) {
	log.Printf("writeMsg - timestamp:%d msg:%v", timestamp, msg)

	doc := client.Doc("Messages/" + msg.ID)
	wr, err := doc.Create(ctx, FsMessage {
		Timestamp:          timestamp,
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
