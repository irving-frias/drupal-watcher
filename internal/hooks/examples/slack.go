package hooks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/irving-frias/drupal-watcher/pkg/core"
)

type SlackNotifier struct {
	WebhookURL string
	client     *http.Client
}

func NewSlackNotifier(webhookURL string) *SlackNotifier {
	return &SlackNotifier{
		WebhookURL: webhookURL,
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *SlackNotifier) Name() string {
	return "SlackNotifier"
}

func (s *SlackNotifier) Process(ctx context.Context, event core.FileEvent, result core.ExecutionResult) error {
	if s.WebhookURL == "" {
		return nil
	}

	status := "success"
	color := "#36a64f"
	if result.ExitCode != 0 {
		status = "failed"
		color = "#ff0000"
	}

	payload := map[string]interface{}{
		"attachments": []map[string]interface{}{
			{
				"color": color,
				"title": "Drupal Watcher: Cache Clear",
				"fields": []map[string]interface{}{
					{"title": "File", "value": event.Path, "short": true},
					{"title": "Command", "value": result.Command, "short": true},
					{"title": "Status", "value": status, "short": true},
					{"title": "Exit Code", "value": fmt.Sprintf("%d", result.ExitCode), "short": true},
					{"title": "Duration", "value": result.Duration.Round(time.Millisecond).String(), "short": true},
				},
				"footer": "drupal-watcher",
				"ts":     time.Now().Unix(),
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("slack marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.WebhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("slack request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("slack post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("slack returned %d", resp.StatusCode)
	}

	return nil
}
