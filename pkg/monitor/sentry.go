package monitor

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"go.uber.org/zap"
)

func checkSentry(logger *zap.Logger, url string, bearerToken string) (bool, string) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, "sentry url misconfigured for monitor"
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", bearerToken))
	res, err := client.Do(req)
	if err != nil {
		return false, "error performing sentry request"
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return false, "failed to read sentry response"
	}

	if res.StatusCode > 299 || res.StatusCode < 200 {
		logger.Warn("bad response from sentry api", zap.String("url", url), zap.String("body", string(body)))
		return false, "bad response from sentry api\n```" + string(body) + "```"
	}

	var issues []Issue
	err = json.Unmarshal(body, &issues)
	if err != nil {
		logger.Warn("bad response from sentry api", zap.String("url", url), zap.String("body", string(body)))
		return false, "bad response from sentry api\n```" + string(body) + "```"
	}

	if len(issues) == 0 {
		return true, ""
	}

	return false, fmt.Sprintf("sentry issue detected: <%s|View in Sentry>", issues[0].WebLink)
}

// Issue is an unmarshal target for a sentry issues request
type Issue struct {
	Title     string    `json:"title,omitempty"`
	FirstSeen time.Time `json:"firstSeen,omitempty"`
	LastSeen  time.Time `json:"lastSeen,omitempty"`
	UserCount int       `json:"userCount,omitempty"`
	WebLink   string    `json:"permalink,omitempty"`
}
