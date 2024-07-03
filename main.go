package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	traq "github.com/traPtitech/go-traq"
	traqwsbot "github.com/traPtitech/traq-ws-bot"
	payload "github.com/traPtitech/traq-ws-bot/payload"
)

func main() {

	// Load the .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Get BOT_TOKEN from environment variables
	token, err := getBotToken()
	if err != nil {
		log.Fatalf("Error getting BOT_TOKEN: %v", err)
	}

	bot, err := traqwsbot.NewBot(&traqwsbot.Options{
		AccessToken: token,
	})
	if err != nil {
		panic(err)
	}

	bot.OnMessageCreated(func(p *payload.MessageCreated) {
		log.Println("Received MESSAGE_CREATED event: " + p.Message.Text)
		_, _, err := bot.API().
			MessageApi.
			PostMessage(context.Background(), p.Message.ChannelID).
			PostMessageRequest(traq.PostMessageRequest{
				Content: "oisu-",
			}).
			Execute()
		if err != nil {
			log.Println(err)
		}
	})

	if err := bot.Start(); err != nil {
		panic(err)
	}
	// Set up HTTP server for WebHook
	http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		var webhookPayload map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&webhookPayload)
		if err != nil {
			http.Error(w, "Error decoding JSON", http.StatusBadRequest)
			return
		}

		log.Println("Received WebHook:", webhookPayload)

		// Process the WebHook payload here according to your needs
		// For example, you can send a message to traQ
		_, _, err = bot.API().
			MessageApi.
			PostMessage(context.Background(), "your_channel_id"). // Specify the correct channel ID
			PostMessageRequest(traq.PostMessageRequest{
				Content: "WebHook received: " + webhookPayload["message"].(string), // Modify as needed
			}).
			Execute()
		if err != nil {
			log.Println(err)
		}

		w.WriteHeader(http.StatusOK)
	})

	go func() {
		log.Println("Starting HTTP server for WebHook on port 8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	if err := bot.Start(); err != nil {
		panic(err)
	}
}

func getBotToken() (string, error) {
	return getEnv("BOT_TOKEN")
}

func getEnv(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", fmt.Errorf("environment variable %s not set", key)
	}
	return value, nil
}
