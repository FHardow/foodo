package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/fhardow/foodo/internal/domain/order"
)

type Notifier struct {
	token  string
	chatID string
	client *http.Client
}

func NewNotifier(token, chatID string) *Notifier {
	return &Notifier{
		token:  token,
		chatID: chatID,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (n *Notifier) OrderConfirmed(o *order.Order) {
	if err := n.send(formatMessage(o)); err != nil {
		slog.Error("telegram notification failed", "err", err, "order_id", o.ID())
	}
}

func (n *Notifier) send(text string) error {
	body, _ := json.Marshal(map[string]string{
		"chat_id": n.chatID,
		"text":    text,
	})
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", n.token)
	resp, err := n.client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API returned %d", resp.StatusCode)
	}
	return nil
}

func formatMessage(o *order.Order) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "New Order\n")
	fmt.Fprintf(&sb, "ID: %s\n", o.ID())
	fmt.Fprintf(&sb, "Customer: %s\n", o.UserName())
	fmt.Fprintf(&sb, "Items (%d):\n", len(o.Items()))
	for _, item := range o.Items() {
		fmt.Fprintf(&sb, "  %dx %s\n", item.Quantity, item.ProductName)
	}
	fmt.Fprintf(&sb, "Total: %.2f\n", float64(o.TotalCents())/100)
	return sb.String()
}
