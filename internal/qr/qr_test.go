package qr

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
)

func qrPNG(t *testing.T, text string) []byte {
	t.Helper()
	matrix, err := qrcode.NewQRCodeWriter().Encode(text, gozxing.BarcodeFormat_QR_CODE, 300, 300, nil)
	if err != nil {
		t.Fatalf("encode QR: %v", err)
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, matrix); err != nil {
		t.Fatalf("encode PNG: %v", err)
	}
	return buf.Bytes()
}

func TestDecodeReadsQR(t *testing.T) {
	const payload = "00460006000001010300201110123456789012345102TH9104ABCD"
	got, ok := Decode(qrPNG(t, payload))
	if !ok || got != payload {
		t.Errorf("Decode = (%q, %v) ต้องการ (%q, true)", got, ok, payload)
	}
}

func TestDecodeWithoutQR(t *testing.T) {
	plain := image.NewRGBA(image.Rect(0, 0, 64, 64))
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			plain.Set(x, y, color.White)
		}
	}
	var buf bytes.Buffer
	_ = png.Encode(&buf, plain)
	if got, ok := Decode(buf.Bytes()); ok {
		t.Errorf("รูปเปล่าต้องไม่เจอ QR ได้ %q", got)
	}
}

func TestDecodeGarbageBytes(t *testing.T) {
	if _, ok := Decode([]byte("not an image")); ok {
		t.Error("ข้อมูลที่ไม่ใช่รูปต้องคืน false ไม่ panic")
	}
	if _, ok := Decode(nil); ok {
		t.Error("nil ต้องคืน false")
	}
}
