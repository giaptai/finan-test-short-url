package bll

import (
	"errors"
	"math/rand"
	"net/url"
	"strings"
	"time"

	"github.com/giaptai/finan-test-short-url/dao"
	"github.com/giaptai/finan-test-short-url/model"
)

const (
	shortCodeLength = 6
	base62Chars     = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	maxRetries      = 5
)

type URLService struct {
	repo *dao.URLRespository
	rnd  *rand.Rand
}

func NewURLService(repo *dao.URLRespository) *URLService {
	return &URLService{
		repo: repo,
		rnd:  rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// create new short link
func (s *URLService) CreateShortURL(originalURL string) (*model.URL, error) {
	// 1 validate url
	if err := s.ValidateURL(originalURL); err != nil {
		return nil, err
	}

	// 2 generate unique short code
	shortCode, err := s.generateUniqueShortCode()
	if err != nil {
		return nil, err
	}

	// create url object
	url := &model.URL{
		ShortCode:   shortCode,
		OriginalURL: originalURL,
		Clicks:      0,
		CreatedAt:   time.Now(),
	}

	// store in database
	if err := s.repo.Add(url); err != nil {
		return nil, err
	}

	return url, nil
}

// get original url
func (s *URLService) GetOriginalURL(shortCode string) (*model.URL, error) {
	url, err := s.repo.GetByShortCode(shortCode)
	if err != nil {
		return nil, err
	}
	if url == nil {
		return nil, errors.New("URL not found")
	}
	return url, nil
}

// increment click
func (s *URLService) IncreaseClick(shortCode string) error {
	return s.repo.IncrementClicks(shortCode)
}

// get all with pagination
func (s *URLService) ListURLs(limit, offset int) ([]*model.URL, error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.GetAll(limit, offset)
}

// validate url
func (s *URLService) ValidateURL(rawURL string) error {
	u, err := url.Parse(rawURL)

	if err != nil {
		return errors.New("Invalid URL format")
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return errors.New("URL must start with http:// or https://")
	}

	if u.Host == "" {
		return errors.New("URL must have a valid host")
	}

	// Blacklist localhost
	if strings.Contains(u.Host, "localhost") ||
		strings.Contains(u.Host, "127.0.0.1") ||
		strings.Contains(u.Host, "0.0.0.0") {
		return errors.New("cannot shorten localhost URLs")
	}

	return nil
}

// generate unique short code
func (s *URLService) generateUniqueShortCode() (string, error) {
	for i := 0; i < maxRetries; i++ {
		shortCode := s.generateRandomCode()

		exists, err := s.repo.ExistsByShortCode(shortCode)
		if err != nil {
			return "", err
		}

		if !exists {
			return shortCode, nil
		}
	}
	return "", errors.New("failed to generate unique short code after retries")
}

// generate random code with 6 characters
func (s *URLService) generateRandomCode() string {
	code := make([]byte, shortCodeLength)
	for i := range code {
		idx := s.rnd.Intn(len(base62Chars))
		code[i] = base62Chars[idx]
	}
	return string(code)
}
