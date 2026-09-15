package slipreader_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	slipreader "github.com/Muegoo/slip-reader"
	"github.com/Muegoo/slip-reader/internal/testfixture"
	"github.com/Muegoo/slip-reader/ocr"
)

func TestReadEveryFixtureThroughPipeline(t *testing.T) {
	for _, c := range testfixture.Load(t) {
		t.Run(c.Name, func(t *testing.T) {
			var want slipreader.Document
			if err := json.Unmarshal(c.WantJSON, &want); err != nil {
				t.Fatalf("JSON เสีย: %v", err)
			}

			reader := slipreader.New(&ocr.Fake{Lines: c.RawLines})
			got, err := reader.Read(context.Background(), []byte("image"))
			if err != nil {
				t.Fatalf("Read: %v", err)
			}

			if got.Type != want.Type || got.Issuer != want.Issuer ||
				got.AmountSatang != want.AmountSatang || got.Counterparty != want.Counterparty ||
				!got.OccurredAt.Equal(want.OccurredAt) || got.Ref != want.Ref || got.Confidence != want.Confidence {
				t.Errorf("Document ไม่ตรง\nได้     %+v\nต้องการ %+v", *got, want)
			}
			if len(got.RawLines) != len(c.RawLines) {
				t.Errorf("RawLines ต้องเก็บข้อความดิบครบ %d บรรทัด ได้ %d", len(c.RawLines), len(got.RawLines))
			}
		})
	}
}

func TestReadPropagatesEngineError(t *testing.T) {
	boom := errors.New("boom")
	doc, err := slipreader.New(&ocr.Fake{Err: boom}).Read(context.Background(), nil)
	if !errors.Is(err, boom) || !strings.Contains(err.Error(), "slipreader") {
		t.Errorf("error = %v ต้องห่อ boom และบอกว่ามาจาก slipreader", err)
	}
	if doc != nil {
		t.Errorf("Document ต้องเป็น nil เมื่อ error ได้ %+v", doc)
	}
}

func TestReadEmptyTextIsNotAnError(t *testing.T) {
	// รูปแมว: OCR คืนอะไรไม่ได้เลย — ไม่ใช่ error แต่เป็นเอกสารที่ทุกฟิลด์ missing
	doc, err := slipreader.New(&ocr.Fake{Lines: nil}).Read(context.Background(), nil)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	all := slipreader.Confidence{Amount: slipreader.LevelMissing, Counterparty: slipreader.LevelMissing, OccurredAt: slipreader.LevelMissing}
	if doc.Issuer != slipreader.IssuerUnknown || doc.Type != slipreader.DocTypeUnknown || doc.Confidence != all {
		t.Errorf("Document = %+v", *doc)
	}
}
