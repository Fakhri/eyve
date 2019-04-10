package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rekognition"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

type Configuration struct {
	Region                 string
	BucketName             string
	CollectionId           string
	FbMessengerAccessToken string
	KnownPersonImageFile   string
}

type GetMessageCreativeIdRequest struct {
	Messages []interface{} `json:"messages"`
}

type SimpleFbMessengerText struct {
	Text string `json:"text"`
	//"attachment":{
	//	"type":"image",
	//	"payload":{
	//		"url":"http://www.messenger-rocks.com/image.jpg",
	//		"is_reusable":true
	//	}
	//}
}

type GetMessageCreativeIdResponse struct {
	MessageCreativeId string `json:"message_creative_id"`
}

type BroadcastMessageRequest struct {
	MessageCreativeId string `json:"message_creative_id"`
	MessagingType     string `json:"messaging_type"`
	Tag               string `json:"tag"`
}

func main() {
	lambda.Start(StartVideoAnalysisHandler)
}

func StartVideoAnalysisHandler(ctx context.Context, event events.S3Event) (string, error) {
	log.Printf("start handling event: %v", event)

	fileName := event.Records[0].S3.Object.Key
	log.Printf("file name: %v", fileName)

	config := Configuration{
		Region:                 os.Getenv("REGION"),
		BucketName:             os.Getenv("S3_BUCKET_NAME"),
		CollectionId:           os.Getenv("COLLECTION_ID"),
		FbMessengerAccessToken: os.Getenv("FB_MESSENGER_ACCESS_TOKEN"),
		KnownPersonImageFile:   os.Getenv("KNOWN_PERSON_IMAGE_FILE"),
	}

	sourceImage := rekognition.Image{
		S3Object: &rekognition.S3Object{
			Bucket: &config.BucketName,
			Name:   &config.KnownPersonImageFile,
		},
	}

	targetImage := rekognition.Image{
		S3Object: &rekognition.S3Object{
			Bucket: &config.BucketName,
			Name:   &fileName,
		},
	}

	startFaceComparison(config, sourceImage, targetImage, fileName)

	return "success", nil
}

func startFaceComparison(config Configuration, sourceImage rekognition.Image, targetImage rekognition.Image, fileName string) {
	session, err := session.NewSession(&aws.Config{Region: aws.String(config.Region)})
	if err != nil {
		log.Fatal(err)
	}

	rek := rekognition.New(session)

	threshold := float64(90)

	input := rekognition.CompareFacesInput{
		SourceImage:         &sourceImage,
		TargetImage:         &targetImage,
		SimilarityThreshold: &threshold,
	}

	output, err := rek.CompareFaces(&input)
	if err != nil {
		log.Fatal(err)
	}

	var strBuilder strings.Builder
	strBuilder.WriteString(fmt.Sprintf("Analysis result for image analysis\n\n"))

	for range output.FaceMatches {
		strBuilder.WriteString(fmt.Sprintf("detected a known face from source: %s, target file: %s\n\n", config.KnownPersonImageFile, fileName))
	}

	foundUnmatched := false

	for _, result := range output.UnmatchedFaces {
		foundUnmatched = true
		confidence := fmt.Sprintf("detected an unknown face with confidence: %f\n\n", *result.Confidence)
		strBuilder.WriteString(confidence)
		log.Println(confidence)
	}

	if foundUnmatched {
		detectedFace := fmt.Sprintf("target image file: https://s3.amazonaws.com/eyve-image-analysis/%s\n", fileName)
		strBuilder.WriteString(detectedFace)
		log.Println(detectedFace)
	}

	log.Println(output)

	message := strBuilder.String()
	log.Println(message)
	broadcastToFbMessenger(config, message)
}

func broadcastToFbMessenger(config Configuration, message string) {
	log.Print("start broadcast message to FB Messenger")

	messageCreativeId, err := getMessageCreativeId(config, message)
	if err != nil {
		log.Fatal(err)
	}

	broadcastMessageUrl := fmt.Sprintf(
		"https://graph.facebook.com/v2.11/me/broadcast_messages?access_token=%s",
		config.FbMessengerAccessToken,
	)

	req := BroadcastMessageRequest{
		MessageCreativeId: messageCreativeId,
		MessagingType:     "MESSAGE_TAG",
		Tag:               "NON_PROMOTIONAL_SUBSCRIPTION",
	}

	log.Println("creative id", messageCreativeId)

	payload, err := json.Marshal(req)
	res, err := http.Post(broadcastMessageUrl, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		log.Fatal(err)
	}

	defer res.Body.Close()

	responseBytes, err := ioutil.ReadAll(res.Body)
	log.Printf("received broadcast response: %s, status code: %d", string(responseBytes), res.StatusCode)

	log.Print("finished broadcast message to FB Messenger")
}

func getMessageCreativeId(config Configuration, message string) (string, error) {
	getMessageCreativeIdUrl := fmt.Sprintf(
		"https://graph.facebook.com/v2.11/me/message_creatives?access_token=%s",
		config.FbMessengerAccessToken,
	)

	simpleText := SimpleFbMessengerText{Text: message}

	messages := []interface{}{simpleText}

	req := GetMessageCreativeIdRequest{Messages: messages}

	payload, err := json.Marshal(req)

	res, err := http.Post(getMessageCreativeIdUrl, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	getMessageCreativeIdResponse := GetMessageCreativeIdResponse{}
	err = json.NewDecoder(res.Body).Decode(&getMessageCreativeIdResponse)
	if err != nil {
		return "", err
	}

	log.Printf(
		"received messageCreativeId: %s, status code: %d",
		getMessageCreativeIdResponse.MessageCreativeId, res.StatusCode,
	)

	return getMessageCreativeIdResponse.MessageCreativeId, nil
}
