package slipreader

import (
	"context"
	"fmt"

	"github.com/Muegoo/slip-reader/internal/extract"
	"github.com/Muegoo/slip-reader/internal/issuer"
	"github.com/Muegoo/slip-reader/internal/qr"
	"github.com/Muegoo/slip-reader/internal/textnorm"
	"github.com/Muegoo/slip-reader/ocr"
)

// New สร้าง Reader ที่ใช้ OCR engine ที่ให้มา (ocr.NewPaddle สำหรับของจริง, ocr.Fake สำหรับเทสต์)
func New(engine ocr.Engine) Reader {
	return &reader{engine: engine}
}

type reader struct {
	engine ocr.Engine
}

// Read ร้อย pipeline: OCR → ทำความสะอาดบรรทัด → เดา issuer → แกะฟิลด์ตามกติกาของ issuer นั้น
func (r *reader) Read(ctx context.Context, image []byte) (*Document, error) {
	raw, err := r.engine.Recognize(ctx, image)
	if err != nil {
		return nil, fmt.Errorf("slipreader: %w", err)
	}

	lines := textnorm.CleanLines(raw)
	who := issuer.Detect(lines)
	result := extract.For(who).Extract(lines)

	// เลขอ้างอิงจาก QR แม่นกว่าจาก OCR (ไม่มีตัวเลขเพี้ยน) จึงใช้แทนถ้าถอดได้
	// ถอดไม่ได้ไม่ใช่ error — ใบเสร็จและ e-wallet บางเจ้าไม่มี QR เลย
	if ref, ok := qr.Decode(image); ok {
		result.Ref = ref
	}
	return buildDocument(who, result, raw), nil
}

// buildDocument แปลง Result ภายในเป็น Document สาธารณะตรง ๆ ไม่มี logic เพิ่ม
// เก็บ raw (ก่อนทำความสะอาด) ไว้ด้วย เพื่อให้แอปรันตัวแกะฟิลด์รุ่นใหม่กับเอกสารเก่าได้โดยไม่ต้อง OCR ซ้ำ
func buildDocument(who string, result extract.Result, raw []string) *Document {
	return &Document{
		Type:         DocType(result.Type),
		Issuer:       Issuer(who),
		AmountSatang: int64(result.AmountSatang),
		OccurredAt:   result.OccurredAt,
		Counterparty: result.Counterparty,
		Ref:          result.Ref,
		Confidence: Confidence{
			Amount:       Level(result.AmountLevel),
			Counterparty: Level(result.CounterpartyLevel),
			OccurredAt:   Level(result.OccurredAtLevel),
		},
		RawLines: raw,
	}
}
