package notifier

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
)

type TelegramNotifier struct {
	Token  string
	ChatID string
}

func NewTelegramNotifier() *TelegramNotifier {
	return &TelegramNotifier{
		Token:  os.Getenv("TELEGRAM_BOT_TOKEN"),
		ChatID: os.Getenv("TELEGRAM_CHAT_ID"),
	}
}

func (t *TelegramNotifier) SendMessage(message string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.Token)
	body := []byte(fmt.Sprintf(`{"chat_id":"%s","text":"%s"}`, t.ChatID, message))

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
