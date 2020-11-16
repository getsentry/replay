package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

/*
https://cloud.google.com/appengine/docs/standard/go111/googlecloudstorageclient/read-write-to-cloud-storage
https://github.com/GoogleCloudPlatform/golang-samples/blob/8deb2909eadf32523007fd8fe9e8755a12c6d463/docs/appengine/storage/app.go
*/
func readJsons(ignore bool) string {
	bucketName := os.Getenv("BUCKET")
	// Initialize/Connect the Client
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalln("storage.NewClient:", err)
	}
	defer client.Close()
	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()
	// Prepare bucket handle
	bucketHandle := client.Bucket(bucketName)
	// lists the contents of a bucket in Google Cloud Storage.
	var fileNames []string
	query := &storage.Query{Prefix: "eventtest"}
	it := bucketHandle.Objects(ctx, query)
	for {
		obj, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalln("listBucket: unable to list bucket", err)
		}
		fileNames = append(fileNames, obj.Name)
		printObj(obj)
	}

	// TODO events.go could manage reading from storage. or like:
	/*
		storageClient := StorageClient(os.Getenv("BUCKET")) <-- is the init
		//or
		storageClient.init(os.Getenv("BUCKET"))
		storageClient.query("event") .prefixQuery("event") .queryBucket .bucketQuery() .bucketSet()
		storageClient.listBucketContents() .getBucket()
		events := storageClient.getFiles() .bucketFiles()
	*/

	// Read each file's content
	var events []EventJson
	for _, fileName := range fileNames {
		rc, err := bucketHandle.Object(fileName).NewReader(ctx)
		if err != nil {
			log.Fatalln("NewReader:", err)
		}
		byteValue, _ := ioutil.ReadAll(rc)
		var event EventJson
		// The EventJson's UnmarshalJSON overriden in event-to-sentry.go (soon EventJson.go)
		if err := json.Unmarshal(byteValue, &event); err != nil {
			panic(err)
		}
		events = append(events, event)
	}

	for _, event := range events {
		if event.Kind == "error" {
			event.Error.eventId()
			event.Error.release()
			event.Error.user()
			event.Error.timestamp()
		}
		if event.Kind == "transaction" {
			event.Transaction.eventId()
			event.Transaction.release()
			event.Transaction.user()
			event.Transaction.timestamps()
		}
	}

	getTraceIds(events)
	updateTraceIds(events)

	// TODO double check it's object was updated reference `fmt.Println("\n> timestamp AFTER", event.Error.Timestamp)`
	requests := Requests{events}
	requests.send()

	return "\n DONE \n"

}

func printObj(obj *storage.ObjectAttrs) {
	fmt.Printf("filename: /%v/%v \n", obj.Bucket, obj.Name)
	// fmt.Printf("ContentType: %q, ", obj.ContentType)
	// fmt.Printf("ACL: %#v, ", obj.ACL)
	// fmt.Printf("Owner: %v, ", obj.Owner)
	// fmt.Printf("ContentEncoding: %q, ", obj.ContentEncoding)
	// fmt.Printf("Size: %v, ", obj.Size)
	// fmt.Printf("MD5: %q, ", obj.MD5)
	// fmt.Printf("CRC32C: %q, ", obj.CRC32C)
	// fmt.Printf("Metadata: %#v, ", obj.Metadata)
	// fmt.Printf("MediaLink: %q, ", obj.MediaLink)
	// fmt.Printf("StorageClass: %q, ", obj.StorageClass)
}