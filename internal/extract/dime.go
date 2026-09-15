package extract

import "github.com/Muegoo/slip-reader/internal/kind"

// dime แกะสลิปโอนเงินจากแอป Dime! (ธ.เกียรตินาคินภัทร)
//
// ครึ่งบนของสลิปเป็นภาพกราฟิกที่ OCR ไม่เห็น บรรทัดแรกที่อ่านได้คือ "โอนเงิน" แล้วตามด้วยยอดทันที
type dime struct{}

func (dime) Extract(lines []string) Result {
	r := emptyResult(kind.DocTransfer)

	if v, ok := amountAfter(lines, "โอนเงิน"); ok {
		r.AmountSatang, r.AmountLevel = v, kind.High
	}
	if t, ok := firstDate(lines); ok {
		r.OccurredAt, r.OccurredAtLevel = t, kind.High
	}
	if i := indexOf(lines, "ไปยัง"); i >= 0 {
		if name, ok := nextTextLine(lines, i); ok {
			r.Counterparty, r.CounterpartyLevel = name, kind.High
		}
	}
	if ref, ok := valueAfterLabel(lines, "เลขที่สลิป"); ok {
		r.Ref = ref
	}
	return r
}
