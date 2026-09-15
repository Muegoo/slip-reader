package ocr

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
)

func TestPaddleSendsImageAndReturnsLines(t *testing.T) {
	image := []byte{0x89, 'P', 'N', 'G', 0, 1, 2}
	var gotBody map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/ocr" {
			t.Errorf("request = %s %s ต้องการ POST /ocr", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decode body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"result":{"ocrResults":[{"prunedResult":{"rec_texts":["โอนเงินสำเร็จ","K+","199.00 บาท"]}}]}}`))
	}))
	defer server.Close()

	lines, err := NewPaddle(server.URL).Recognize(context.Background(), image)
	if err != nil {
		t.Fatalf("Recognize: %v", err)
	}
	want := []string{"โอนเงินสำเร็จ", "K+", "199.00 บาท"}
	if !slices.Equal(lines, want) {
		t.Errorf("lines = %q ต้องการ %q", lines, want)
	}

	decoded, err := base64.StdEncoding.DecodeString(gotBody["file"].(string))
	if err != nil || !slices.Equal(decoded, image) {
		t.Errorf("file ที่ส่งไปไม่ใช่รูปเดิม (err=%v)", err)
	}
	if gotBody["fileType"] != float64(1) {
		t.Errorf("fileType = %v ต้องการ 1", gotBody["fileType"])
	}
}

func TestPaddleReportsServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "model exploded", http.StatusInternalServerError)
	}))
	defer server.Close()

	_, err := NewPaddle(server.URL).Recognize(context.Background(), []byte("x"))
	if err == nil || !strings.Contains(err.Error(), "500") || !strings.Contains(err.Error(), "model exploded") {
		t.Errorf("error = %v ต้องมีทั้งรหัส 500 และข้อความจาก server", err)
	}
}

func TestPaddleHonoursContextCancel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done() // ค้างจนกว่า client จะยกเลิก
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := NewPaddle(server.URL).Recognize(ctx, []byte("x"))
	if !errors.Is(err, context.Canceled) {
		t.Errorf("error = %v ต้องเป็น context.Canceled", err)
	}
}

func TestFake(t *testing.T) {
	lines, err := (&Fake{Lines: []string{"a"}}).Recognize(context.Background(), nil)
	if err != nil || !slices.Equal(lines, []string{"a"}) {
		t.Errorf("Fake = %q, %v", lines, err)
	}
	boom := errors.New("boom")
	if _, err := (&Fake{Err: boom}).Recognize(context.Background(), nil); !errors.Is(err, boom) {
		t.Errorf("Fake error = %v", err)
	}
}
