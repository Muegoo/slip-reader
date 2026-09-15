package ocr

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Paddle เรียก ocr-service (PaddleOCR หลัง paddlex serving) ผ่าน HTTP
type Paddle struct {
	BaseURL string
	Client  *http.Client
}

// สลิปใบละ ~3 วินาที แต่รูปจากกล้อง 13MB อาจนานกว่านั้นมาก จึงให้เผื่อไว้พอ
const defaultTimeout = 60 * time.Second

// NewPaddle สร้าง client ที่ชี้ไป base URL ของ ocr-service เช่น http://localhost:8866
func NewPaddle(baseURL string) *Paddle {
	return &Paddle{
		BaseURL: strings.TrimRight(baseURL, "/"),
		Client:  &http.Client{Timeout: defaultTimeout},
	}
}

// รูปแบบ request/response ของ paddlex serving pipeline OCR
// (ยืนยันกับ container จริงใน ocr-service/README.md — ถ้า upstream เปลี่ยน ให้แก้ตรงนี้ที่เดียว)
type paddleRequest struct {
	File     string `json:"file"`     // รูปเป็น base64
	FileType int    `json:"fileType"` // 1 = รูปภาพ (0 = PDF)
}

type paddleResponse struct {
	Result struct {
		OCRResults []struct {
			PrunedResult struct {
				RecTexts []string `json:"rec_texts"`
			} `json:"prunedResult"`
		} `json:"ocrResults"`
	} `json:"result"`
}

func (p *Paddle) Recognize(ctx context.Context, image []byte) ([]string, error) {
	body, err := json.Marshal(paddleRequest{
		File:     base64.StdEncoding.EncodeToString(image),
		FileType: 1,
	})
	if err != nil {
		return nil, fmt.Errorf("ocr: encode request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.BaseURL+"/ocr", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("ocr: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ocr: call %s: %w", p.BaseURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// เอาข้อความจาก server ติดมาด้วย (ตัดให้สั้น) จะได้รู้ว่าพังเพราะอะไร ไม่ใช่แค่รหัส
		snippet, _ := io.ReadAll(io.LimitReader(resp.Body, 200))
		return nil, fmt.Errorf("ocr: server returned %d: %s", resp.StatusCode, strings.TrimSpace(string(snippet)))
	}

	var parsed paddleResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("ocr: decode response: %w", err)
	}

	var lines []string
	for _, page := range parsed.Result.OCRResults {
		lines = append(lines, page.PrunedResult.RecTexts...)
	}
	return lines, nil
}
