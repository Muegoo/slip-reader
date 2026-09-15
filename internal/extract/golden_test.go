package extract

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/Muegoo/slip-reader/internal/issuer"
	"github.com/Muegoo/slip-reader/internal/testfixture"
	"github.com/Muegoo/slip-reader/internal/textnorm"
)

// expected สะท้อนรูปแบบ JSON ใน testdata/ocr/*/*.json (ตรงกับ slipreader.Document แต่ไม่ import เพื่อเลี่ยง cycle)
type expected struct {
	Type         string    `json:"type"`
	Issuer       string    `json:"issuer"`
	AmountSatang int64     `json:"amount_satang"`
	OccurredAt   time.Time `json:"occurred_at"`
	Counterparty string    `json:"counterparty"`
	Ref          string    `json:"ref"`
	Confidence   struct {
		Amount       string `json:"amount"`
		Counterparty string `json:"counterparty"`
		OccurredAt   string `json:"occurred_at"`
	} `json:"confidence"`
}

func TestGolden(t *testing.T) {
	for _, c := range testfixture.Load(t) {
		t.Run(c.Name, func(t *testing.T) {
			var want expected
			if err := json.Unmarshal(c.WantJSON, &want); err != nil {
				t.Fatalf("JSON ของ %s เสีย: %v", c.Name, err)
			}

			lines := textnorm.CleanLines(c.RawLines)
			who := issuer.Detect(lines)
			got := For(who).Extract(lines)

			check := func(field string, got, want any) {
				if got != want {
					t.Errorf("%s: ได้ %v ต้องการ %v", field, got, want)
				}
			}
			check("type", got.Type, want.Type)
			check("amount_satang", int64(got.AmountSatang), want.AmountSatang)
			check("confidence.amount", got.AmountLevel, want.Confidence.Amount)
			check("counterparty", got.Counterparty, want.Counterparty)
			check("confidence.counterparty", got.CounterpartyLevel, want.Confidence.Counterparty)
			check("occurred_at", got.OccurredAt.Equal(want.OccurredAt), true)
			check("confidence.occurred_at", got.OccurredAtLevel, want.Confidence.OccurredAt)
			check("ref", got.Ref, want.Ref)
		})
	}
}
