package extract

import (
	"strings"

	"github.com/Muegoo/slip-reader/internal/kind"
)

// sevenEleven แกะใบเสร็จจาก 7App (ฟอนต์เลียนเครื่องพิมพ์ความร้อน)
//
// ใบเดียวมี ยอดรวม / ส่วนลดหลายบรรทัด / ยอดสุทธิ — ต้องเอา "ยอดสุทธิ" และบางใบไม่มีบรรทัด "ยอดรวม" เลย
type sevenEleven struct{}

func (sevenEleven) Extract(lines []string) Result {
	r := emptyResult(kind.DocReceipt)

	// "ยอดสุทธิ 2 ชั้น 41.50" — จำนวนชิ้นอยู่ระหว่างคำนำหน้ากับยอด (บางใบแยกบรรทัด บางใบไม่)
	// ตัดบรรทัดจำนวนชิ้นออกก่อน ไม่งั้นจะได้ 2.00 บาท
	priced := withoutItemCounts(lines)
	if v, ok := amountAfter(priced, "ยอดสุทธิ"); ok {
		r.AmountSatang, r.AmountLevel = v, kind.High
	} else if v, ok := amountAfter(priced, "7App"); ok {
		// บรรทัด "ทรูวอลเล็ท7App 41.50" คือยอดที่จ่ายจริงเช่นกัน แต่ OCR อ่านคำหน้าเพี้ยนบ่อย จึงมั่นใจน้อยกว่า
		r.AmountSatang, r.AmountLevel = v, kind.Low
	}
	if t, ok := firstDate(lines); ok {
		r.OccurredAt, r.OccurredAtLevel = t, kind.High
	}
	if i := indexOf(lines, "สาขา"); i >= 0 {
		r.Counterparty, r.CounterpartyLevel = branchName(lines[i]), kind.High
	}
	if i := indexOf(lines, "TID#"); i >= 0 {
		r.Ref = strings.TrimPrefix(lines[i], "TID#")
	}
	return r
}

func withoutItemCounts(lines []string) []string {
	kept := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.Contains(line, "ชั้น") || strings.Contains(line, "ชิ้น") {
			continue
		}
		kept = append(kept, line)
	}
	return kept
}

// branchName ตัดคำว่า "สาขา" นำหน้าออก และแก้ชื่อแบรนด์ที่ OCR อ่าน n เป็น ท ("7-Eleveท")
// แก้เฉพาะชื่อแบรนด์ที่รู้แน่ ๆ ไม่แก้ชื่อสาขาซึ่งเราไม่รู้ว่าที่ถูกคืออะไร
func branchName(line string) string {
	name := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "สาขา"))
	return strings.ReplaceAll(name, "7-Eleveท", "7-Eleven")
}
