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
	if i := indexOf(lines, "G-Wallet ID"); i >= 0 {
		if name, ok := paotangShopName(lines, i); ok {
			r.Counterparty, r.CounterpartyLevel = name, kind.High
		}
	}
	if ref, ok := valueAfterLabel(lines, "รหัสอ้างอิง"); ok && paotangRefPattern.MatchString(ref) {
		r.Ref = ref
	}
	return r
}

// บรรทัดที่บอกว่าชื่อร้านจบแล้ว: โลโก้ถุงเงิน (OCR อ่านเป็น "ถูงเอ็น") หรือบรรทัดหมวดหมู่ร้าน
var paotangShopNameStops = []string{"ถุงเงิน", "ถูงเอ็น", "อาหาร", "ของหวาน", "เครื่องดื่ม", "ค่าสินค้า"}

// paotangShopName ต่อบรรทัดข้อความติดกันหลังบล็อกผู้จ่าย (G-Wallet ID + เลขท้าย) เป็นชื่อร้าน
// เพราะ OCR แบ่งชื่อยาวเป็นหลายบรรทัด ("Maki" / "แซลมอน", "รุ่มรวยก๋วยเตี๋ยวไก่" / "สาขา เมกะบางนา")
func paotangShopName(lines []string, walletIndex int) (string, bool) {
	start := nextTextLineIndex(lines, walletIndex)
	if start < 0 {
		return "", false
	}
	var parts []string
	for _, line := range lines[start:] {
		if !isText(line) || containsAny(line, paotangShopNameStops) {
			break
		}
		parts = append(parts, line)
	}
	if len(parts) == 0 {
		return "", false
	}
	return strings.Join(parts, " "), true
}

func containsAny(line string, words []string) bool {
	for _, w := range words {
		if strings.Contains(line, w) {
			return true
		}
	}
	return false
}
