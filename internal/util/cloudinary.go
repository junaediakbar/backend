package util

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type cloudinaryUploadResponse struct {
	SecureURL string `json:"secure_url"`
	URL       string `json:"url"`
	Error     any    `json:"error"`
}

// UploadOrderImageToCloudinary uploads one image; slotIndex distinguishes multiple photos per order (0, 1, 2).
func UploadOrderImageToCloudinary(ctx context.Context, orderID string, slotIndex int, imageBytes []byte) (string, error) {
	cloudinaryURL := strings.TrimSpace(os.Getenv("CLOUDINARY_URL"))
	if cloudinaryURL == "" {
		return "", errors.New("CLOUDINARY_URL is not set")
	}

	u, err := url.Parse(cloudinaryURL)
	if err != nil {
		return "", fmt.Errorf("invalid CLOUDINARY_URL: %w", err)
	}
	if u.User == nil {
		return "", errors.New("invalid CLOUDINARY_URL: missing credentials")
	}
	if u.Host == "" {
		return "", errors.New("invalid CLOUDINARY_URL: missing cloud name")
	}

	apiKey := u.User.Username()
	apiSecret, _ := u.User.Password()
	cloudName := u.Host
	if apiKey == "" || apiSecret == "" {
		return "", errors.New("invalid CLOUDINARY_URL: missing api key/secret")
	}

	endpoint := fmt.Sprintf("https://api.cloudinary.com/v1_1/%s/image/upload", cloudName)

	// NOTE: we pass `public_id` without extension. Cloudinary will store it as an image.
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	part, err := writer.CreateFormFile("file", "upload.jpg")
	if err != nil {
		return "", fmt.Errorf("multipart create file part failed: %w", err)
	}
	if _, err := part.Write(imageBytes); err != nil {
		return "", fmt.Errorf("multipart write file failed: %w", err)
	}

	publicID := orderID
	if slotIndex >= 0 {
		publicID = fmt.Sprintf("%s_%d", orderID, slotIndex)
	}
	_ = writer.WriteField("public_id", publicID)
	_ = writer.WriteField("folder", "orders")
	_ = writer.WriteField("overwrite", "true")

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("multipart close failed: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, &body)
	if err != nil {
		return "", fmt.Errorf("create http request failed: %w", err)
	}
	req.SetBasicAuth(apiKey, apiSecret)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 60 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("cloudinary upload request failed: %w", err)
	}
	defer res.Body.Close()

	respBody, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("cloudinary upload read response failed: %w", err)
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return "", fmt.Errorf("cloudinary upload failed status=%d body=%s", res.StatusCode, string(respBody))
	}

	var out cloudinaryUploadResponse
	if err := json.Unmarshal(respBody, &out); err != nil {
		return "", fmt.Errorf("cloudinary upload invalid json: %w", err)
	}
	if out.SecureURL != "" {
		if len(out.SecureURL) > 80 {
			log.Printf("cloudinary_upload ok public_id=%s secure_url_prefix=%s", publicID, out.SecureURL[:80])
		} else {
			log.Printf("cloudinary_upload ok public_id=%s secure_url=%s", publicID, out.SecureURL)
		}
		return out.SecureURL, nil
	}
	if out.URL != "" {
		if len(out.URL) > 80 {
			log.Printf("cloudinary_upload ok public_id=%s url_prefix=%s", publicID, out.URL[:80])
		} else {
			log.Printf("cloudinary_upload ok public_id=%s url=%s", publicID, out.URL)
		}
		return out.URL, nil
	}

	return "", fmt.Errorf("cloudinary upload missing secure_url: %s", string(respBody))
}

// uploadOrderImageToCloudinaryWithPublicID uploads with an explicit public_id (unique per attachment).
func uploadOrderImageToCloudinaryWithPublicID(ctx context.Context, publicID string, imageBytes []byte) (string, error) {
	cloudinaryURL := strings.TrimSpace(os.Getenv("CLOUDINARY_URL"))
	if cloudinaryURL == "" {
		return "", errors.New("CLOUDINARY_URL is not set")
	}

	u, err := url.Parse(cloudinaryURL)
	if err != nil {
		return "", fmt.Errorf("invalid CLOUDINARY_URL: %w", err)
	}
	if u.User == nil {
		return "", errors.New("invalid CLOUDINARY_URL: missing credentials")
	}
	if u.Host == "" {
		return "", errors.New("invalid CLOUDINARY_URL: missing cloud name")
	}

	apiKey := u.User.Username()
	apiSecret, _ := u.User.Password()
	cloudName := u.Host
	if apiKey == "" || apiSecret == "" {
		return "", errors.New("invalid CLOUDINARY_URL: missing api key/secret")
	}

	endpoint := fmt.Sprintf("https://api.cloudinary.com/v1_1/%s/image/upload", cloudName)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	part, err := writer.CreateFormFile("file", "upload.jpg")
	if err != nil {
		return "", fmt.Errorf("multipart create file part failed: %w", err)
	}
	if _, err := part.Write(imageBytes); err != nil {
		return "", fmt.Errorf("multipart write file failed: %w", err)
	}

	_ = writer.WriteField("public_id", publicID)
	_ = writer.WriteField("folder", "orders")
	_ = writer.WriteField("overwrite", "true")

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("multipart close failed: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, &body)
	if err != nil {
		return "", fmt.Errorf("create http request failed: %w", err)
	}
	req.SetBasicAuth(apiKey, apiSecret)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 60 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("cloudinary upload request failed: %w", err)
	}
	defer res.Body.Close()

	respBody, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("cloudinary upload read response failed: %w", err)
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return "", fmt.Errorf("cloudinary upload failed status=%d body=%s", res.StatusCode, string(respBody))
	}

	var out cloudinaryUploadResponse
	if err := json.Unmarshal(respBody, &out); err != nil {
		return "", fmt.Errorf("cloudinary upload invalid json: %w", err)
	}
	if out.SecureURL != "" {
		return out.SecureURL, nil
	}
	if out.URL != "" {
		return out.URL, nil
	}

	return "", fmt.Errorf("cloudinary upload missing secure_url: %s", string(respBody))
}

// UploadOrderAttachmentImageToCloudinary uploads a documentation photo; public_id is unique per call.
// seq distinguishes multiple files in one request (same nanosecond).
func UploadOrderAttachmentImageToCloudinary(ctx context.Context, orderID string, seq int, imageBytes []byte) (string, error) {
	orderID = strings.TrimSpace(orderID)
	if orderID == "" {
		return "", errors.New("order id is empty")
	}
	publicID := fmt.Sprintf("%s_att_%s_%d", orderID, strconv.FormatInt(time.Now().UnixNano(), 10), seq)
	return uploadOrderImageToCloudinaryWithPublicID(ctx, publicID, imageBytes)
}

