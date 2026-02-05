package github

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"ai-code-reviewer/internal/config"
	"ai-code-reviewer/internal/observability"

	"github.com/golang-jwt/jwt/v4"
)

type Client interface {
	GetPRDiff(ctx context.Context, repo string, pr int) (string, error)
}

type client struct {
	cfg    *config.Config
	logger *observability.Logger
	http   *http.Client
}

func NewClient(cfg *config.Config, logger *observability.Logger) Client {
	return &client{
		cfg:    cfg,
		logger: logger,
		http:   &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *client) createJWT() (string, error) {

	key, err := loadPrivateKey(c.cfg.GithubPrivateKeyPath)
	if err != nil {
		return "", err
	}

	now := time.Now()

	claims := jwt.RegisteredClaims{
		IssuedAt:  jwt.NewNumericDate(now.Add(-1 * time.Minute)),
		ExpiresAt: jwt.NewNumericDate(now.Add(9 * time.Minute)),
		Issuer:    c.cfg.GithubAppID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	return token.SignedString(key)
}

func loadPrivateKey(path string) (*rsa.PrivateKey, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read key: %w", err)
	}

	block, _ := pem.Decode(b)
	if block == nil {
		return nil, fmt.Errorf("invalid pem")
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err == nil {
		return key, nil
	}

	pkcs8, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return pkcs8.(*rsa.PrivateKey), nil
}

func (c *client) installationToken(ctx context.Context) (string, error) {

	jwt, err := c.createJWT()
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf(
		"https://api.github.com/app/installations/%s/access_tokens",
		c.cfg.GithubInstallationID,
	)

	req, _ := http.NewRequestWithContext(ctx, "POST", url, nil)
	req.Header.Set("Authorization", "Bearer "+jwt)
	req.Header.Set("Accept", "application/vnd.github+json")

	res, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	var r struct {
		Token string `json:"token"`
	}

	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return "", err
	}

	return r.Token, nil
}

func (c *client) GetPRDiff(ctx context.Context, repo string, pr int) (string, error) {

	token, err := c.installationToken(ctx)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf(
		"https://api.github.com/repos/%s/pulls/%d",
		repo, pr,
	)

	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github.diff")

	res, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return "", fmt.Errorf("github status: %d", res.StatusCode)
	}

	b, _ := io.ReadAll(res.Body)

	return string(b), nil
}
