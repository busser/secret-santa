package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const ovhEndpoint = "https://eu.api.ovh.com/1.0"

// ovhClient calls the OVHcloud API, signed with the credentials of an
// application created at https://eu.api.ovh.com/createToken/.
type ovhClient struct {
	applicationKey    string
	applicationSecret string
	consumerKey       string
}

func ovhClientFromEnv() (ovhClient, error) {
	c := ovhClient{
		applicationKey:    os.Getenv("OVH_APPLICATION_KEY"),
		applicationSecret: os.Getenv("OVH_APPLICATION_SECRET"),
		consumerKey:       os.Getenv("OVH_CONSUMER_KEY"),
	}
	if c.applicationKey == "" || c.applicationSecret == "" || c.consumerKey == "" {
		return ovhClient{}, fmt.Errorf("OVH_APPLICATION_KEY, OVH_APPLICATION_SECRET and OVH_CONSUMER_KEY must be set")
	}
	return c, nil
}

// call sends a request to path, relative to ovhEndpoint, and decodes the JSON
// response into result, unless result is nil.
func (c ovhClient) call(method, path string, body, result interface{}) error {
	var payload []byte
	if body != nil {
		var err error
		if payload, err = json.Marshal(body); err != nil {
			return fmt.Errorf("encoding request body: %w", err)
		}
	}

	url := ovhEndpoint + path
	req, err := http.NewRequest(method, url, bytes.NewReader(payload))
	if err != nil {
		return err
	}

	// OVH rejects timestamps that differ from its clock by more than a few
	// seconds. We trust the local clock, which macOS keeps in sync.
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	toSign := strings.Join([]string{c.applicationSecret, c.consumerKey, method, url, string(payload), timestamp}, "+")
	signature := sha1.Sum([]byte(toSign))

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Ovh-Application", c.applicationKey)
	req.Header.Set("X-Ovh-Consumer", c.consumerKey)
	req.Header.Set("X-Ovh-Timestamp", timestamp)
	req.Header.Set("X-Ovh-Signature", "$1$"+hex.EncodeToString(signature[:]))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s %s: %s: %s", method, path, resp.Status, respBody)
	}
	if result == nil {
		return nil
	}
	if err := json.Unmarshal(respBody, result); err != nil {
		return fmt.Errorf("decoding response: %w", err)
	}
	return nil
}

// sendSMS sends one non-commercial text from the SMS account service, using
// OVH's short number as the sender. It returns the number of credits used.
func (c ovhClient) sendSMS(service, phoneNumber, message string) (credits float64, err error) {
	request := map[string]interface{}{
		"receivers":         []string{phoneNumber},
		"message":           message,
		"charset":           "UTF-8",
		"noStopClause":      true,
		"senderForResponse": true,
		"priority":          "high",
	}
	var report struct {
		InvalidReceivers    []string `json:"invalidReceivers"`
		ValidReceivers      []string `json:"validReceivers"`
		TotalCreditsRemoved float64  `json:"totalCreditsRemoved"`
	}
	if err := c.call(http.MethodPost, "/sms/"+service+"/jobs", request, &report); err != nil {
		return 0, err
	}
	if len(report.ValidReceivers) != 1 {
		return 0, fmt.Errorf("OVH did not accept the phone number (invalid receivers: %d)", len(report.InvalidReceivers))
	}
	return report.TotalCreditsRemoved, nil
}
