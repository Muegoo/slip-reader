package extract

import "github.com/Muegoo/slip-reader/internal/kind"

// bbl แกะสลิปโอนเงินจากแอป Bangkok Bank
//
// layout นิ่งที่สุดในชุดข้อมูล: จำนวนเงิน / ยอด / จาก / ผู้ส่ง / ไปที่ / ผู้รับ / ค่าธรรมเนียม / เลขที่อ้างอิง
type bbl struct{}

func (bbl) Extract(lines []string) Result {
	r := emptyResult(kind.DocTransfer)

	if v, ok := amountAfter(lines, "จำนวนเงิน"); ok {
		r.AmountSatang, r.AmountLevel = v, kind.High
	}
	if t, ok := firstDate(lines); ok {
		r.OccurredAt, r.OccurredAtLevel = t, kind.High
	}
	if i := indexOf(lines, "ไปที่"); i >= 0 {
		if name, ok := nextTextLine(lines, i); ok {
			r.Counterparty, r.CounterpartyLevel = name, kind.High
		}
	}
	// "เลขที่อ้างอิง" (ยาว 25 หลัก) ไม่ใช่ "หมายเลขอ้างอิง" (6 หลัก) — ตัวยาวคือกุญแจกันบันทึกซ้ำที่ดีกว่า
	if ref, ok := valueAfterLabel(lines, "เลขที่อ้างอิง"); ok {
		r.Ref = ref
	}
	return r
}
