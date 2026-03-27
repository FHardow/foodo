package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/fhardow/bread-order/internal/domain/order"
)

// Notifier sends Telegram messages via the Bot API.
type Notifier struct {
	botToken string
	chatID   string
	client   *http.Client
}

// New creates a Notifier. Both botToken and chatID must be non-empty.
func New(botToken, chatID string) *Notifier {
	return &Notifier{
		botToken: botToken,
		chatID:   chatID,
		client:   &http.Client{Timeout: 10 * time.Second},
	}
}

// OrderCreated implements order.Notifier.
func (n *Notifier) OrderCreated(ctx context.Context, o *order.Order) error {
	text := fmt.Sprintf(
		"🛒 New order created\nID: %s\nUser: %s\nStatus: %s\nCreated: %s",
		o.ID(),
		o.UserID(),
		o.Status(),
		o.CreatedAt().Format(time.RFC3339),
	)
	return n.sendMessage(ctx, text)
}

func (n *Notifier) sendMessage(ctx context.Context, text string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", n.botToken)

	body, err := json.Marshal(map[string]string{
		"chat_id": n.chatID,
		"text":    text,
	})
	if err != nil {
		return fmt.Errorf("telegram: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("telegram: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("telegram: send message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram: unexpected status %d", resp.StatusCode)
	}
	return nil
}
