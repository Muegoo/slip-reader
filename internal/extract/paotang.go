package extract

import (
	"regexp"
	"strings"

	"github.com/Muegoo/slip-reader/internal/kind"
)

// paotang แกะใบเสร็จร้านค้าจากแอปเป๋าตัง (โครงการไทยช่วยไทย)
//
// ใบเดียวมีสามยอด: ค่าสินค้า / สิทธิส่วนลด (ติดลบ) / จำนวนเงินที่ชำระ — ต้องเอาตัวสุดท้ายเท่านั้น
// วันเวลาอาจแยกกัน 2-3 บรรทัด และยอดอาจแยกจากคำว่า "บาท" — helper ร่วมรองรับทั้งคู่แล้ว
type paotang struct{}

// รหัสอ้างอิงเป็น hex 32 ตัว หรือ UUID
var paotangRefPattern = regexp.MustCompile(`^[0-9a-f]{32}$|^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

func (paotang) Extract(lines []string) Result {
	r := emptyResult(kind.DocReceipt)

	if v, ok := amountAfter(lines, "จำนวนเงินที่ชำระ"); ok {
		r.AmountSatang, r.AmountLevel = v, kind.High
	}
	if t, ok := firstDate(lines); ok {
		r.OccurredAt, r.OccurredAtLevel = t, kind.High
	}
	// ชื่อร้านคือบรรทัดข้อความแรกหลังบล็อกผู้จ่าย (G-Wallet ID + เลขท้าย)
	// ถ้าชื่อยาว แอปตัดคำว่า "สาขา …" ลงบรรทัดถัดไป — ต่อกลับให้ครบ เพราะสาขาเป็นส่วนของชื่อร้าน
	if i := indexOf(lines, "G-Wallet ID"); i >= 0 {
		if j := nextTextLineIndex(lines, i); j >= 0 {
			name := lines[j]
			if j+1 < len(lines) && strings.HasPrefix(lines[j+1], "สาขา") {
				name += " " + lines[j+1]
			}
			r.Counterparty, r.CounterpartyLevel = name, kind.High
		}
	}
	if ref, ok := valueAfterLabel(lines, "รหัสอ้างอิง"); ok && paotangRefPattern.MatchString(ref) {
		r.Ref = ref
	}
	return r
}
