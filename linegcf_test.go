package linegcf

import (
	"reflect"
	"testing"
)

func TestWriteMsg(t *testing.T) {
	testID := "testID"
	var (
		testTimestamp int64 = 123456789
		testFileSize int64 = 12345
		testDuration int64 = 12345
	)
	var testData = LineMessage {
		ID:              testID,
		Type:            "test Type",
		Text:            "test Text",
		PackageID:       "test PackageID",
		StickerID:       "test StickerID",
		FileName:        "test FileName",
		FileSize:        testFileSize,
		Title:           "test Title",
		Address:         "test Address",
		Latitude:        1234.5,
		Longitude:       1234.5,
		Duration:        testDuration,
		ContentProvider: struct {
			Type               string `json: "type"`
			OriginalContentURL string `json: "originalContentUrl"`
			PreviewImageURL    string `json: "previewImageUrl"`
		}{
			Type:               "test ContentType",
			OriginalContentURL: "test OriginalContentURL",
			PreviewImageURL:    "test PreviewImageURL",
		},
	}
	expected := map[string]interface{} {
		"Timestamp":          testTimestamp,
		"Type":               "test Type",
		"Text":               "test Text",
		"PackageID":          "test PackageID",
		"StickerID":          "test StickerID",
		"FileName":           "test FileName",
		"FileSize":           testFileSize,
		"Title":              "test Title",
		"Address":            "test Address",
		"Latitude":           1234.5,
		"Longitude":          1234.5,
		"Duration":           testDuration,
		"ContentType":        "test ContentType",
		"OriginalContentURL": "test OriginalContentURL",
		"PreviewImageURL":    "test PreviewImageURL",
	}

	writeMsg(testTimestamp, testData)

	docRef := client.Doc("Messages/" + testID)
	docsnap, err := docRef.Get(ctx)
	if err != nil {
		t.Errorf("firestore docRef.Get: %v", err)
	}
	r := docsnap.Data()

	if !reflect.DeepEqual(r, expected) {
		docRef.Delete(ctx)
		t.Errorf("WriteResult does not deep equal expected: %v", r)
	}

	docRef.Delete(ctx)
}
