package extract

import "github.com/Muegoo/slip-reader/internal/kind"

// generic คือตัวสำรองเมื่อไม่รู้จัก issuer — อ่านได้เท่าที่อ่านได้แล้วให้ผู้ใช้เติม ดีกว่าโยน error ทิ้ง
//
// ทุกอย่างที่หาเจอได้ level low เพราะไม่รู้ layout จึงไม่กล้ามั่นใจ
// และไม่พยายามเดาชื่อคู่ค้าเลย — ชื่อผิดแล้วบันทึกไปเงียบ ๆ แย่กว่าให้ผู้ใช้พิมพ์เอง
type generic struct{}

// คำนำหน้ายอดที่พบบนสลิปไทย เรียงจากเฉพาะเจาะจงไปทั่วไป — "จำนวน" ต้องมาหลัง "จำนวนเงินที่ชำระ"
// ไม่งั้นจะไปจับบรรทัด "จำนวนเงินที่ชำระ" แล้วอ่านผิดความหมาย
var genericAmountLabels = []string{"จำนวนเงินที่ชำระ", "ยอดสุทธิ", "จำนวนเงิน", "จำนวน", "ยอดรวม"}

func (generic) Extract(lines []string) Result {
	r := emptyResult(kind.DocUnknown)

	for _, label := range genericAmountLabels {
		if v, ok := amountAfter(lines, label); ok {
			r.AmountSatang, r.AmountLevel = v, kind.Low
			break
		}
	}
	if t, ok := firstDate(lines); ok {
		r.OccurredAt, r.OccurredAtLevel = t, kind.Low
	}
	return r
}
