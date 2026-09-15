package extract

import (
	"regexp"

	"github.com/Muegoo/slip-reader/internal/kind"
)

// kbank แกะสลิปจากแอป K PLUS
//
// layout: หัว / K+ / วันเวลา / ชื่อผู้ส่ง / ธ.กสิกรไทย / เลขบัญชีผู้ส่ง / ชื่อผู้รับ / ... / เลขที่รายการ: / จำนวน: / ค่าธรรมเนียม:
// ทุกใบเป็นเงินออกจากบัญชี ไม่ว่าหัวจะเขียน โอน/ชำระ/จ่ายบิล/เติมเงิน จึงเป็น transfer ทั้งหมด
type kbank struct{}

// เลขที่รายการของ K PLUS เช่น 672878907298DQR65814 หรือตัวเลขล้วน 18 หลัก
var kbankRefPattern = regexp.MustCompile(`^[0-9A-Z]{15,}$`)

func (kbank) Extract(lines []string) Result {
	r := emptyResult(kind.DocTransfer)

	if v, ok := amountAfter(lines, "จำนวน"); ok {
		r.AmountSatang, r.AmountLevel = v, kind.High
	}
	if t, ok := firstDate(lines); ok {
		r.OccurredAt, r.OccurredAtLevel = t, kind.High
	}
	r.Counterparty, r.CounterpartyLevel = kbankCounterparty(lines)
	if ref, ok := valueAfterLabel(lines, "เลขที่รายการ"); ok && kbankRefPattern.MatchString(ref) {
		r.Ref = ref
	}
	return r
}

// kbankCounterparty: ผู้รับคือบรรทัดข้อความแรกหลังเลขบัญชีของผู้ส่ง (บล็อกบนคือผู้ส่งเสมอ)
// ถ้า OCR อ่านเลขบัญชีไม่ออก ใช้บรรทัดหลัง "ธ.กสิกรไทย" ตัวแรกแทน แต่มั่นใจน้อยกว่า
func kbankCounterparty(lines []string) (string, string) {
	for i, line := range lines {
		if looksLikeAccount(line) {
			if name, ok := nextTextLine(lines, i); ok {
				return name, kind.High
			}
			break
		}
	}
	if i := indexOf(lines, "ธ.กสิกรไทย"); i >= 0 {
		if name, ok := nextTextLine(lines, i); ok {
			return name, kind.Low
		}
	}
	return "", kind.Missing
}
