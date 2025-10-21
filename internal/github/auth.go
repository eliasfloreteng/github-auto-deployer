package github

import (
	"context"
	"crypto/rsa"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/go-github/v57/github"
)

// AppClient represents a GitHub App client
type AppClient struct {
	appID          int64
	installationID int64
	privateKey     *rsa.PrivateKey
	client         *github.Client
}

// NewAppClient creates a new GitHub App client
func NewAppClient(appID, installationID int64, privateKeyPath string) (*AppClient, error) {
	keyData, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %w", err)
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(keyData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	ac := &AppClient{
		appID:          appID,
		installationID: installationID,
		privateKey:     privateKey,
	}

	// Create initial client with installation token
	if err := ac.refreshClient(); err != nil {
		return nil, fmt.Errorf("failed to create initial client: %w", err)
	}

	return ac, nil
}

// refreshClient creates a new GitHub client with a fresh installation token
func (ac *AppClient) refreshClient() error {
	// Create JWT for GitHub App authentication
	now := time.Now()
	claims := jwt.RegisteredClaims{
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(10 * time.Minute)),
		Issuer:    fmt.Sprintf("%d", ac.appID),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signedToken, err := token.SignedString(ac.privateKey)
	if err != nil {
		return fmt.Errorf("failed to sign JWT: %w", err)
	}

	// Create temporary client with JWT
	jwtClient := github.NewClient(nil).WithAuthToken(signedToken)

	// Get installation token
	ctx := context.Background()
	installToken, _, err := jwtClient.Apps.CreateInstallationToken(
		ctx,
		ac.installationID,
		&github.InstallationTokenOptions{},
	)
	if err != nil {
		return fmt.Errorf("failed to create installation token: %w", err)
	}

	// Create authenticated client with installation token
	ac.client = github.NewClient(nil).WithAuthToken(installToken.GetToken())

	return nil
}

// GetClient returns the GitHub client
func (ac *AppClient) GetClient() *github.Client {
	return ac.client
}

// ValidateWebhookSignature validates the webhook signature
func ValidateWebhookSignature(payload []byte, signature string, secret string) bool {
	err := github.ValidateSignature(signature, payload, []byte(secret))
	return err == nil
}

// ParseWebhookPayload parses a webhook payload
func ParseWebhookPayload(r *http.Request) (interface{}, error) {
	payload, err := github.ValidatePayload(r, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to validate payload: %w", err)
	}

	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		return nil, fmt.Errorf("failed to parse webhook: %w", err)
	}

	return event, nil
}
