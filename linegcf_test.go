package linegcf

import (
	"log"
	"reflect"
	"testing"
)

func TestWriteMsg(t *testing.T) {
	testID := "testID"
	var (
		testTimestamp int64 = 123456789
		testFileSize  int64 = 12345
		testDuration  int64 = 12345
	)
	var testData = LineMessage{
		ID:        testID,
		Type:      "test Type",
		Text:      "test Text",
		PackageID: "test PackageID",
		StickerID: "test StickerID",
		FileName:  "test FileName",
		FileSize:  testFileSize,
		Title:     "test Title",
		Address:   "test Address",
		Latitude:  1234.5,
		Longitude: 1234.5,
		Duration:  testDuration,
		ContentProvider: struct {
			Type               string `json:"type"`
			OriginalContentURL string `json:"originalContentUrl"`
			PreviewImageURL    string `json:"previewImageUrl"`
		}{
			Type:               "test ContentType",
			OriginalContentURL: "test OriginalContentURL",
			PreviewImageURL:    "test PreviewImageURL",
		},
	}
	var testSource = LineSource{
		Type:    "test SourceType",
		UserID:  "test UserID",
		GroupID: "test GroupID",
		RoomID:  "test RoomID",
	}
	expected := map[string]interface{}{
		"timestamp":          testTimestamp,
		"replyToken":         "test ReplyToken",
		"sourceType":         "test SourceType",
		"userID":             "test UserID",
		"groupID":            "test GroupID",
		"roomID":             "test RoomID",
		"type":               "test Type",
		"text":               "test Text",
		"packageID":          "test PackageID",
		"stickerID":          "test StickerID",
		"fileName":           "test FileName",
		"fileSize":           testFileSize,
		"title":              "test Title",
		"address":            "test Address",
		"latitude":           1234.5,
		"longitude":          1234.5,
		"duration":           testDuration,
		"contentType":        "test ContentType",
		"originalContentURL": "test OriginalContentURL",
		"previewImageURL":    "test PreviewImageURL",
	}

	writeMsg(testTimestamp, "test ReplyToken", testSource, testData)

	docRef := client.Doc("Messages/" + testID)
	docsnap, err := docRef.Get(ctx)
	if err != nil {
		t.Errorf("firestore docRef.Get: %v", err)
	}
	r := docsnap.Data()

	log.Printf("result:%v", r["sourceType"])
	log.Printf("expected:%v", expected["sourceType"])
	log.Printf("result:%v", r["userID"])
	log.Printf("expected:%v", expected["userID"])
	log.Printf("result:%v", r["groupID"])
	log.Printf("expected:%v", expected["groupID"])

	defer docRef.Delete(ctx)
	if !reflect.DeepEqual(r, expected) {
		t.Errorf("WriteResult does not deep equal expected: %v", r)
	}
}
