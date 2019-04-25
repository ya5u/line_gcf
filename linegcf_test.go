package linegcf

import (
	"log"
	"reflect"
	"testing"
)

func TestWriteMsg(t *testing.T) {
	const testID = "9876543210"
	var	testTimestamp int64 = 123456789
	var testData = TextMessage{
		testTimestamp,
		"test ReplyToken",
		"test SourceType",
		"test UserID",
		"test GroupID",
		"test RoomID",
		"test Type",
		"test Text",
	}
	expected := map[string]interface{}{
		"timestamp":  testTimestamp,
		"replyToken": "test ReplyToken",
		"sourceType": "test SourceType",
		"userID":     "test UserID",
		"groupID":    "test GroupID",
		"roomID":     "test RoomID",
		"type":       "test Type",
		"text":       "test Text",
	}

	wg.Add(1)
	go writeMsg(testID, testData)
	wg.Wait()

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
