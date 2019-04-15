package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

var FbMessengerAccessToken = os.Getenv("FB_MESSENGER_ACCESS_TOKEN")

type GetMessageCreativeIdRequest struct {
	Messages []interface{} `json:"messages"`
}

type SimpleFbMessengerText struct {
	Text string `json:"text"`
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
	message := "Stella triggered the panic alert (with voice command). \n" +
		"Location : Singapore Expo" +
		"Please call the nearest police station: Changi Neighbourhood Police Centre (6587 2999)"

	broadcastToFbMessenger(message)
}

func broadcastToFbMessenger(message string) {
	log.Print("start broadcast message to FB Messenger")

	messageCreativeId, err := getMessageCreativeId(message)
	if err != nil {
		log.Fatal(err)
	}

	broadcastMessageUrl := fmt.Sprintf(
		"https://graph.facebook.com/v2.11/me/broadcast_messages?access_token=%s",
		FbMessengerAccessToken,
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

func getMessageCreativeId(message string) (string, error) {
	getMessageCreativeIdUrl := fmt.Sprintf(
		"https://graph.facebook.com/v2.11/me/message_creatives?access_token=%s",
		FbMessengerAccessToken,
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
