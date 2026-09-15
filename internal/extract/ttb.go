package extract

import (
	"strings"

	"github.com/Muegoo/slip-reader/internal/amount"
	"github.com/Muegoo/slip-reader/internal/kind"
)

// ttb แกะสลิปจ่ายบิลจากแอป ttb touch
//
// ยอดเป็นตัวเลขใหญ่ลอย ๆ ไม่มีคำนำหน้า — เป็น issuer เดียวที่ต้องพึ่งตำแหน่ง:
// ยอดคือตัวเลขเงินตัวแรกหลังบรรทัดวันเวลา และก่อนบรรทัด "ค่าธรรมเนียม"
// โลโก้ ttb ถูกอ่านเป็นบรรทัด "ttb" แทรกอยู่ทั่วสลิป ต้องข้ามเวลาหาชื่อผู้รับ
type ttb struct{}

func (ttb) Extract(lines []string) Result {
	r := emptyResult(kind.DocTransfer)

	t, dateIndex, hasDate := firstDateSpan(lines)
	if hasDate {
		r.OccurredAt, r.OccurredAtLevel = t, kind.High
	}
	if v, ok := ttbAmount(lines, dateIndex); ok {
		r.AmountSatang, r.AmountLevel = v, kind.High
	} else if v, ok := amountBeforeFee(lines); ok {
		// อ่านวันที่ไม่ออก (OCR อ่าน "มิ.ย." เป็น "ม.ย.") ไม่ควรลากยอดให้หายไปด้วย —
		// ยอดของ ttb อยู่บรรทัดก่อน "ค่าธรรมเนียม" เสมอ แต่ไม่มีวันที่ยึดจึงมั่นใจน้อยกว่า
		r.AmountSatang, r.AmountLevel = v, kind.Low
	}
	if name, ok := ttbCounterparty(lines); ok {
		r.Counterparty, r.CounterpartyLevel = name, kind.High
	}
	if ref, ok := valueAfterLabel(lines, "รหัสอ้างอิง"); ok {
		r.Ref = ref
	}
	return r
}

func ttbAmount(lines []string, dateIndex int) (amount.Satang, bool) {
	if dateIndex < 0 {
		return 0, false
	}
	for _, line := range lines[dateIndex+1:] {
		if strings.Contains(line, "ค่าธรรมเนียม") {
			return 0, false
		}
		if v, ok := amount.Parse(line); ok {
			return v, true
		}
	}
	return 0, false
}

// amountBeforeFee คืนยอดจากบรรทัดที่อยู่ก่อนบรรทัด "ค่าธรรมเนียม" ทันที
func amountBeforeFee(lines []string) (amount.Satang, bool) {
	i := indexOf(lines, "ค่าธรรมเนียม")
	if i <= 0 {
		return 0, false
	}
	return amount.Parse(lines[i-1])
}

// ttbCounterparty: ผู้รับบิลคือบรรทัดข้อความแรกหลังเลขบัญชี/บัตรของผู้จ่าย โดยข้ามบรรทัดโลโก้ "ttb"
func ttbCounterparty(lines []string) (string, bool) {
	for i, line := range lines {
		if !looksLikeAccount(line) {
			continue
		}
		for _, candidate := range lines[i+1:] {
			if isTTBLogo(candidate) || !isText(candidate) {
				continue
			}
			return candidate, true
		}
		return "", false
	}
	return "", false
}

func isTTBLogo(line string) bool {
	lower := strings.ToLower(line)
	return lower == "ttb" || lower == "utb"
}
