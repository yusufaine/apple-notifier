package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/yusufaine/apple-inventory-notifier/pkg/rlclient"
	"golang.org/x/time/rate"
)

type ParseMode string

const (
	ParseHTML       ParseMode = "html"
	ParseMarkdown   ParseMode = "Markdown"
	ParseMarkdownV2 ParseMode = "MarkdownV2"
)

type Bot struct {
	botEp  url.URL
	chatId string
	ctx    context.Context
	rlc    *rlclient.Client
}

func NewBot(c *Config) *Bot {
	ep := url.URL{
		Scheme: "https",
		Host:   "api.telegram.org",
		Path:   "/bot" + c.Token,
	}

	// 1 request per 4 seconds -- tg's limit is 20 qpm, conservatively set to 15 qpm
	rl := rate.NewLimiter(rate.Every(4*time.Second), 1)
	tg := &Bot{
		botEp:  ep,
		chatId: c.ChatId,
		ctx:    c.Context,
		rlc:    rlclient.New(c.Context, rlclient.WithRateLimiter(rl)),
	}

	return tg
}

func (b *Bot) Delete(msgId int) error {
	reqBody, err := json.Marshal(map[string]interface{}{
		"chat_id":    b.chatId,
		"message_id": msgId,
	})
	if err != nil {
		return err
	}

	ep := b.botEp
	ep.Path += "/deleteMessage"
	req, err := newTgRequest(b.ctx, http.MethodPost, ep.String(), bytes.NewReader(reqBody))
	if err != nil {
		return err
	}

	_, err = b.rlc.Do(req)
	if err != nil {
		return err
	}

	return nil
}

func (b *Bot) Send(text string, p ParseMode) (int, error) {
	reqBody, err := json.Marshal(map[string]interface{}{
		"chat_id":                  b.chatId,
		"disable_web_page_preview": true,
		"parse_mode":               string(p),
		"text":                     text,
	})
	if err != nil {
		return 0, err
	}

	ep := b.botEp
	ep.Path += "/sendMessage"
	req, err := newTgRequest(b.ctx, http.MethodPost, ep.String(), bytes.NewReader(reqBody))
	if err != nil {
		return 0, err
	}

	body, err := b.rlc.Do(req)
	if err != nil {
		return 0, err
	}

	var resp struct {
		Result struct {
			MessageID int `json:"message_id"`
		} `json:"result"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return 0, err
	}

	return resp.Result.MessageID, nil
}

// This blocks the program for the specified duration
func (b *Bot) TempSend(text string, d time.Duration, p ParseMode) {
	msgId, err := b.Send(text, p)
	if err != nil {
		slog.Error("unable to temp write to telegram", slog.String("error", err.Error()))
		return
	}
	time.Sleep(d)
	b.Delete(msgId)
}

// Convenience method to always set the content type of the request
func newTgRequest(ctx context.Context, method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}
