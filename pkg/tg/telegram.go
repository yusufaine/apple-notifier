package tg

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/time/rate"
)

type ParseMode string

const (
	ParseHTML       ParseMode = "html"
	ParseMarkdown   ParseMode = "Markdown"
	ParseMarkdownV2 ParseMode = "Markdown"
)

type Bot struct {
	botEp  url.URL
	chatId string
	ctx    context.Context
	hc     *http.Client
	rl     *rate.Limiter
}

func NewBot(c *Config) *Bot {
	ep := url.URL{
		Scheme: "https",
		Host:   "api.telegram.org",
		Path:   "/bot" + c.Token,
	}

	tg := &Bot{
		botEp:  ep,
		chatId: c.ChatId,
		ctx:    c.Context,
		hc:     http.DefaultClient, // potentially make a client wrapper with rate limiting
		rl:     rate.NewLimiter(rate.Every(time.Minute), 15),
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
	req, err := b.newRequestWithHeader(http.MethodPost, ep.String(), bytes.NewReader(reqBody))
	if err != nil {
		return err
	}

	_, err = b.doRequest(req)
	if err != nil {
		return err
	}

	return nil
}

func (tw *Bot) Write(text string, p ParseMode) (int, error) {
	reqBody, err := json.Marshal(map[string]interface{}{
		"chat_id":                  tw.chatId,
		"disable_web_page_preview": true,
		"parse_mode":               string(p),
		"text":                     text,
	})
	if err != nil {
		return 0, err
	}

	ep := tw.botEp
	ep.Path += "/sendMessage"
	req, err := tw.newRequestWithHeader(http.MethodPost, ep.String(), bytes.NewReader(reqBody))
	if err != nil {
		return 0, err
	}

	body, err := tw.doRequest(req)
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
func (tw *Bot) TempWrite(text string, d time.Duration, p ParseMode) {
	msgId, err := tw.Write(text, p)
	if err != nil {
		slog.Error("unable to temp write to telegram", slog.String("error", err.Error()))
		return
	}
	time.Sleep(d)
	tw.Delete(msgId)
}

func (b *Bot) doRequest(req *http.Request) ([]byte, error) {
	if err := b.rl.Wait(b.ctx); err != nil {
		return nil, err
	}

	resp, err := b.hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf(string(body))
	}

	return body, nil
}

// Convenience method to always set the content type of the request
func (tw *Bot) newRequestWithHeader(method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(tw.ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}
