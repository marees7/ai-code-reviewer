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

type client struct {
	cfg    *config.Config
	logger *observability.Logger
	http   *http.Client
	cache  *tokenCache
}

func NewClient(cfg *config.Config, logger *observability.Logger) Client {
	return &client{
		cfg:    cfg,
		logger: logger,
		http:   &http.Client{Timeout: 15 * time.Second},
		cache:  &tokenCache{},
	}
}

func (c *client) getToken(ctx context.Context) (string, error) {

	if t, ok := c.cache.Get(); ok {
		return t, nil
	}

	jwt, err := c.createJWT()
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf(
		"https://api.github.com/app/installations/%s/access_tokens",
		c.cfg.GithubInstallationID,
	)

	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return "", fmt.Errorf("build token request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+jwt)
	req.Header.Set("Accept", "application/vnd.github+json")

	res, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		msg, _ := io.ReadAll(io.LimitReader(res.Body, 4096))
		return "", fmt.Errorf("github token status %d: %s", res.StatusCode, string(msg))
	}

	var r struct {
		Token     string `json:"token"`
		ExpiresAt string `json:"expires_at"`
	}

	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return "", fmt.Errorf("decode token response: %w", err)
	}
	if r.Token == "" {
		return "", fmt.Errorf("empty installation token")
	}

	c.cache.Set(r.Token, 50*time.Minute)

	return r.Token, nil
}

func (c *client) GetPRFiles(ctx context.Context, repo string, pr int) ([]PRFile, error) {

	var files []PRFile

	err := withRetry(3, func() error {

		token, err := c.getToken(ctx)
		if err != nil {
			return err
		}

		url := fmt.Sprintf(
			"https://api.github.com/repos/%s/pulls/%d/files",
			repo, pr,
		)

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return fmt.Errorf("build files request: %w", err)
		}

		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Accept", "application/vnd.github+json")

		res, err := c.http.Do(req)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		if res.StatusCode == 403 {
			return fmt.Errorf("rate limited")
		}
		if res.StatusCode < 200 || res.StatusCode >= 300 {
			msg, _ := io.ReadAll(io.LimitReader(res.Body, 4096))
			return fmt.Errorf("github files status %d: %s", res.StatusCode, string(msg))
		}

		b, err := io.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("read files response: %w", err)
		}

		if err := json.Unmarshal(b, &files); err != nil {
			return fmt.Errorf("decode files response: %w", err)
		}
		return nil
	})

	// Filter only reviewable files
	var result []PRFile
	for _, f := range files {
		if IsReviewable(f) {
			result = append(result, f)
		}
	}

	c.logger.Info("files fetched",
		"total", len(files),
		"reviewable", len(result),
	)

	return result, err
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

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("build diff request: %w", err)
	}

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

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("read diff response: %w", err)
	}

	return string(b), nil
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

	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return "", fmt.Errorf("build installation token request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+jwt)
	req.Header.Set("Accept", "application/vnd.github+json")

	res, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		msg, _ := io.ReadAll(io.LimitReader(res.Body, 4096))
		return "", fmt.Errorf("github installation token status %d: %s", res.StatusCode, string(msg))
	}

	var r struct {
		Token string `json:"token"`
	}

	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return "", err
	}

	return r.Token, nil
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

	privateKey, ok := pkcs8.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("pkcs8 key is not RSA")
	}

	return privateKey, nil
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

// create comment
func (c *client) CreateComment(ctx context.Context, repo string, pr int, body string) error {
	// Todo: implement this method to create a comment on the PR using GitHub API
	return nil
}
