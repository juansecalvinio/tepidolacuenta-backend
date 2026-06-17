package mercadopago

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/mercadopago/sdk-go/pkg/config"
	"github.com/mercadopago/sdk-go/pkg/payment"
	"github.com/mercadopago/sdk-go/pkg/preference"
)

// PreferenceRequest contains the data needed to create a MercadoPago preference
type PreferenceRequest struct {
	Title             string
	Quantity          int
	UnitPrice         float64
	ExternalReference string
	NotificationURL   string
	BackURLSuccess    string
	BackURLFailure    string
	BackURLPending    string
}

// PreferenceResponse contains the result of creating a preference
type PreferenceResponse struct {
	ID        string
	InitPoint string
}

// PaymentInfo contains the relevant fields from a MercadoPago payment
type PaymentInfo struct {
	ID                int64
	Status            string
	ExternalReference string
}

// Client wraps the MercadoPago SDK
type Client struct {
	cfg           *config.Config
	webhookSecret string
}

func NewClient(accessToken, webhookSecret string) (*Client, error) {
	cfg, err := config.New(accessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create MercadoPago config: %w", err)
	}

	return &Client{
		cfg:           cfg,
		webhookSecret: webhookSecret,
	}, nil
}

// CreatePreference creates a payment preference in MercadoPago and returns the checkout URL
func (c *Client) CreatePreference(ctx context.Context, req PreferenceRequest) (*PreferenceResponse, error) {
	client := preference.NewClient(c.cfg)

	request := preference.Request{
		Items: []preference.ItemRequest{
			{
				Title:     req.Title,
				Quantity:  req.Quantity,
				UnitPrice: req.UnitPrice,
			},
		},
		BackURLs: &preference.BackURLsRequest{
			Success: req.BackURLSuccess,
			Failure: req.BackURLFailure,
			Pending: req.BackURLPending,
		},
		NotificationURL:   req.NotificationURL,
		ExternalReference: req.ExternalReference,
	}

	// MercadoPago solo permite auto_return con back_urls públicas: si se manda
	// con una URL local, rechaza la creación de la preferencia. Por eso lo
	// activamos únicamente cuando la URL de retorno no es localhost.
	if !strings.Contains(req.BackURLSuccess, "localhost") &&
		!strings.Contains(req.BackURLSuccess, "127.0.0.1") {
		request.AutoReturn = "approved"
	}

	result, err := client.Create(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to create MercadoPago preference: %w", err)
	}

	return &PreferenceResponse{
		ID:        result.ID,
		InitPoint: result.InitPoint,
	}, nil
}

// GetPayment retrieves payment details from MercadoPago by numeric payment ID
func (c *Client) GetPayment(ctx context.Context, paymentID int64) (*PaymentInfo, error) {
	client := payment.NewClient(c.cfg)

	result, err := client.Get(ctx, int(paymentID))
	if err != nil {
		return nil, fmt.Errorf("failed to get MercadoPago payment %d: %w", paymentID, err)
	}

	return &PaymentInfo{
		ID:                int64(result.ID),
		Status:            result.Status,
		ExternalReference: result.ExternalReference,
	}, nil
}

// ValidateWebhookSignature validates the x-signature header sent by MercadoPago.
// Returns true if the signature is valid or if no webhook secret is configured (dev mode).
// Signature format: ts=TIMESTAMP,v1=HMAC_SHA256
// Manifest: id:{paymentID};request-id:{xRequestID};ts:{timestamp}
func (c *Client) ValidateWebhookSignature(xSignature, xRequestID, paymentID string) bool {
	if c.webhookSecret == "" {
		return true
	}

	// El formato IPN legacy (topic=payment) no envía x-signature: no hay HMAC que
	// validar. Se confía en la verificación posterior contra la API de MP
	// (GetPayment con el access token propio) dentro de ProcessPaymentWebhook.
	if xSignature == "" {
		return true
	}

	ts, v1 := parseSignatureHeader(xSignature)
	if ts == "" || v1 == "" {
		return false
	}

	// Formato exacto del manifest de MercadoPago (incluye el ; final; el id va en
	// minúsculas si fuera alfanumérico).
	manifest := fmt.Sprintf("id:%s;request-id:%s;ts:%s;", strings.ToLower(paymentID), xRequestID, ts)

	mac := hmac.New(sha256.New, []byte(c.webhookSecret))
	mac.Write([]byte(manifest))
	expected := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(expected), []byte(v1))
}

// parseSignatureHeader parses "ts=...,v1=..." and returns (ts, v1)
func parseSignatureHeader(header string) (ts, v1 string) {
	for _, part := range strings.Split(header, ",") {
		kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(kv) != 2 {
			continue
		}
		switch kv[0] {
		case "ts":
			ts = kv[1]
		case "v1":
			v1 = kv[1]
		}
	}
	return
}
