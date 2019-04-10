package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rekognition"
	"log"
	"os"
)

type Configuration struct {
	Region     string
	BucketName string
}

func main() {
	lambda.Start(TextDetectionHandler)
}

func TextDetectionHandler(ctx context.Context, event events.S3Event) (string, error) {
	log.Printf("start handling event: %v", event)

	fileName := event.Records[0].S3.Object.Key
	log.Printf("S3 file name: %v", fileName)

	config := Configuration{
		Region:     os.Getenv("REGION"),
		BucketName: os.Getenv("S3_BUCKET_NAME"),
	}

	session, err := session.NewSession(&aws.Config{Region: aws.String(config.Region)})
	if err != nil {
		return "", err
	}

	rek := rekognition.New(session)

	img := rekognition.Image{
		S3Object: &rekognition.S3Object{
			Bucket: &config.BucketName,
			Name:   &fileName,
		},
	}

	input := rekognition.DetectTextInput{
		Image: &img,
	}

	output, err := rek.DetectText(&input)
	if err != nil {
		return "", err
	}

	for _, textDetection := range output.TextDetections {
		log.Printf("detected text: %s", *textDetection.DetectedText)
		log.Printf("confidence: %f", *textDetection.Confidence)
	}

	return "ok", nil
}
