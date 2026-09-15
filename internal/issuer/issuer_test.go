package issuer

import (
	"testing"

	"github.com/Muegoo/slip-reader/internal/kind"
	"github.com/Muegoo/slip-reader/internal/testfixture"
	"github.com/Muegoo/slip-reader/internal/textnorm"
)

func TestDetectEveryFixture(t *testing.T) {
	for _, c := range testfixture.Load(t) {
		t.Run(c.Name, func(t *testing.T) {
			got := Detect(textnorm.CleanLines(c.RawLines))
			if got != c.Issuer {
				t.Errorf("Detect = %q ต้องการ %q", got, c.Issuer)
			}
		})
	}
}

func TestDetectPrefersSenderOverRecipientBank(t *testing.T) {
	// สลิป K PLUS ที่โอนไป Dime มีชื่อธนาคารของ Dime อยู่ในช่องผู้รับ — ต้องตอบ kbank ไม่ใช่ dime
	lines := []string{"โอนเงินสำเร็จ", "K+", "1 ก.ย. 69 07:35 น.", "นาย สมชาย ใ", "ธ.กสิกรไทย",
		"xxx-x-x1234-x", "นาย สมชาย ใจดี", "ธ.เกียรตินาคินภัทร", "xxx-x-x0640-x"}
	if got := Detect(lines); got != kind.KBank {
		t.Errorf("Detect = %q ต้องการ kbank", got)
	}

	// สลิป K PLUS ที่โอนไป ttb ก็เช่นกัน
	lines = []string{"โอนเงินสำเร็จ", "K+", "นาย สมชาย ใ", "ธ.กสิกรไทย", "นาง สมหญิง รักดี", "ธ.ทหารไทยธนชาต"}
	if got := Detect(lines); got != kind.KBank {
		t.Errorf("Detect = %q ต้องการ kbank", got)
	}
}

func TestDetectEmpty(t *testing.T) {
	if got := Detect(nil); got != kind.Unknown {
		t.Errorf("Detect(nil) = %q ต้องการ unknown", got)
	}
}
