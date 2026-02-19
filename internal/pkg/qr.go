package pkg

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// QRService handles QR code generation
type QRService struct {
	baseURL string
}

// NewQRService creates a new QR service
func NewQRService(baseURL string) *QRService {
	return &QRService{
		baseURL: baseURL,
	}
}

// GenerateTableQRCode generates a unique QR code for a table
// The QR code contains a URL that points to the request page with encrypted data
func (s *QRService) GenerateTableQRCode(restaurantID, branchID, tableID primitive.ObjectID, tableNumber int) string {
	// Create a payload with restaurant, branch and table info
	payload := fmt.Sprintf("%s:%s:%s:%d", restaurantID.Hex(), branchID.Hex(), tableID.Hex(), tableNumber)

	// Generate a hash for security (could be enhanced with encryption)
	hash := sha256.Sum256([]byte(payload))
	encodedPayload := base64.URLEncoding.EncodeToString(hash[:])

	// Create the QR code URL
	// Format: https://tepidolacuenta.com/request?r=restaurantID&b=branchID&t=tableID&n=number&h=hash
	qrCodeURL := fmt.Sprintf("%s/request?r=%s&b=%s&t=%s&n=%d&h=%s",
		s.baseURL,
		restaurantID.Hex(),
		branchID.Hex(),
		tableID.Hex(),
		tableNumber,
		encodedPayload[:16], // Use first 16 chars of hash
	)

	return qrCodeURL
}

// ValidateTableQRCode validates that a QR code is authentic
func (s *QRService) ValidateTableQRCode(restaurantID, branchID, tableID primitive.ObjectID, tableNumber int, hash string) bool {
	// Generate expected hash
	payload := fmt.Sprintf("%s:%s:%s:%d", restaurantID.Hex(), branchID.Hex(), tableID.Hex(), tableNumber)
	expectedHash := sha256.Sum256([]byte(payload))
	expectedEncoded := base64.URLEncoding.EncodeToString(expectedHash[:])

	// Compare first 16 chars
	return expectedEncoded[:16] == hash
}
