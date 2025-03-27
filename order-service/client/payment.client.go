package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"order-service/client/request"
	"time"
)

type PaymentClientInterface interface {
	Charge(payload request.PaymentRequest) (int, error)
}

type PaymentClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewPaymentClient(baseURL string) *PaymentClient {
	return &PaymentClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (wc *PaymentClient) Charge(payload request.PaymentRequest) (int, error) {
	url := fmt.Sprintf("%s/payment/charge", wc.BaseURL)

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := wc.HTTPClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	return resp.StatusCode, nil
}
