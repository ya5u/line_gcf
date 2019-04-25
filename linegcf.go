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
	"sync"

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
	wg                sync.WaitGroup
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

// TextMessage Document in Firestore collection "Messages"
type TextMessage struct {
	Timestamp  int64  `firestore:"timestamp"`
	ReplyToken string `firestore:"replyToken"`
	SourceType string `firestore:"sourceType"`
	UserID     string `firestore:"userID"`
	GroupID    string `firestore:"groupID"`
	RoomID     string `firestore:"roomID"`
	Type       string `firestore:"type"`
	Text       string `firestore:"text"`
}

// ImageMessage Document in Firestore collection "Messages"
type ImageMessage struct {
	Timestamp          int64  `firestore:"timestamp"`
	ReplyToken         string `firestore:"replyToken"`
	SourceType         string `firestore:"sourceType"`
	UserID             string `firestore:"userID"`
	GroupID            string `firestore:"groupID"`
	RoomID             string `firestore:"roomID"`
	Type               string `firestore:"type"`
	ContentType        string `firestore:"contentType"`
	OriginalContentURL string `firestore:"originalContentURL"`
	PreviewImageURL    string `firestore:"previewImageURL"`
}

// VideoMessage Document in Firestore collection "Messages"
type VideoMessage struct {
	Timestamp          int64  `firestore:"timestamp"`
	ReplyToken         string `firestore:"replyToken"`
	SourceType         string `firestore:"sourceType"`
	UserID             string `firestore:"userID"`
	GroupID            string `firestore:"groupID"`
	RoomID             string `firestore:"roomID"`
	Type               string `firestore:"type"`
	ContentType        string `firestore:"contentType"`
	OriginalContentURL string `firestore:"originalContentURL"`
	PreviewImageURL    string `firestore:"previewImageURL"`
	Duration           int64  `firestore:"duration"`
}

// FileMessage Document in Firestore collection "Messages"
type FileMessage struct {
	Timestamp  int64  `firestore:"timestamp"`
	ReplyToken string `firestore:"replyToken"`
	SourceType string `firestore:"sourceType"`
	UserID     string `firestore:"userID"`
	GroupID    string `firestore:"groupID"`
	RoomID     string `firestore:"roomID"`
	Type       string `firestore:"type"`
	FileName   string `firestore:"fileName"`
	FileSize   int64  `firestore:"fileSize"`
}

// LocationMessage Document in Firestore collection "Messages"
type LocationMessage struct {
	Timestamp  int64   `firestore:"timestamp"`
	ReplyToken string  `firestore:"replyToken"`
	SourceType string  `firestore:"sourceType"`
	UserID     string  `firestore:"userID"`
	GroupID    string  `firestore:"groupID"`
	RoomID     string  `firestore:"roomID"`
	Type       string  `firestore:"type"`
	Title      string  `firestore:"title"`
	Address    string  `firestore:"address"`
	Latitude   float64 `firestore:"latitude"`
	Longitude  float64 `firestore:"longitude"`
}

// StickerMessage Document in Firestore collection "Messages"
type StickerMessage struct {
	Timestamp  int64  `firestore:"timestamp"`
	ReplyToken string `firestore:"replyToken"`
	SourceType string `firestore:"sourceType"`
	UserID     string `firestore:"userID"`
	GroupID    string `firestore:"groupID"`
	RoomID     string `firestore:"roomID"`
	Type       string `firestore:"type"`
	PackageID  string `firestore:"packageID"`
	StickerID  string `firestore:"stickerID"`
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

	events := req.Events

	for _, e := range events {
		wg.Add(1)
		switch e.Message.Type {
		case "text":
			go writeMsg(e.Message.ID, TextMessage{
				e.Timestamp,
				e.ReplyToken,
				e.Source.Type,
				e.Source.UserID,
				e.Source.GroupID,
				e.Source.RoomID,
				e.Message.Type,
				e.Message.Text,
			})
		case "image":
			go writeMsg(e.Message.ID, ImageMessage{
				e.Timestamp,
				e.ReplyToken,
				e.Source.Type,
				e.Source.UserID,
				e.Source.GroupID,
				e.Source.RoomID,
				e.Message.Type,
				e.Message.ContentProvider.Type,
				e.Message.ContentProvider.OriginalContentURL,
				e.Message.ContentProvider.PreviewImageURL,
			})
		case "video":
			fallthrough
		case "audio":
			go writeMsg(e.Message.ID, VideoMessage{
				e.Timestamp,
				e.ReplyToken,
				e.Source.Type,
				e.Source.UserID,
				e.Source.GroupID,
				e.Source.RoomID,
				e.Message.Type,
				e.Message.ContentProvider.Type,
				e.Message.ContentProvider.OriginalContentURL,
				e.Message.ContentProvider.PreviewImageURL,
				e.Message.Duration,
			})
		case "file":
			go writeMsg(e.Message.ID, FileMessage{
				e.Timestamp,
				e.ReplyToken,
				e.Source.Type,
				e.Source.UserID,
				e.Source.GroupID,
				e.Source.RoomID,
				e.Message.Type,
				e.Message.FileName,
				e.Message.FileSize,
			})
		case "location":
			go writeMsg(e.Message.ID, LocationMessage{
				e.Timestamp,
				e.ReplyToken,
				e.Source.Type,
				e.Source.UserID,
				e.Source.GroupID,
				e.Source.RoomID,
				e.Message.Type,
				e.Message.Title,
				e.Message.Address,
				e.Message.Latitude,
				e.Message.Longitude,
			})
		case "sticker":
			go writeMsg(e.Message.ID, StickerMessage{
				e.Timestamp,
				e.ReplyToken,
				e.Source.Type,
				e.Source.UserID,
				e.Source.GroupID,
				e.Source.RoomID,
				e.Message.Type,
				e.Message.PackageID,
				e.Message.StickerID,
			})
		default:
			log.Printf("Type: %s", e.Message.Type)
		}
	}

	wg.Wait()

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

func writeMsg(id string, msg interface{}) {
	log.Printf("writeMsg - id:%s msg:%v", id, msg)

	defer wg.Done()

	doc := client.Doc("Messages/" + id)
	wr, err := doc.Create(ctx, msg)
	if err != nil {
		log.Fatalf("writeMsg: %v", err)
	}
	log.Printf("WriteResult: %v", wr)
}
