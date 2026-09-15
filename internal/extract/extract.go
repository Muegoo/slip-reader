// Package extract แกะฟิลด์ (ยอด คู่ค้า เวลา เลขอ้างอิง) ออกจากบรรทัดข้อความของสลิป
//
// แต่ละ issuer มีไฟล์ของตัวเองที่บอกว่าฟิลด์อยู่ตรงไหน โดยใช้ helper ร่วมในไฟล์นี้
// package นี้รู้จักแค่ []string — ไม่รู้ว่าข้อความมาจาก OCR ตัวไหน
package extract

import (
	"regexp"
	"strings"
	"time"

	"github.com/Muegoo/slip-reader/internal/amount"
	"github.com/Muegoo/slip-reader/internal/kind"
	"github.com/Muegoo/slip-reader/internal/textnorm"
	"github.com/Muegoo/slip-reader/internal/thaidate"
)

// Result คือฟิลด์ที่แกะได้ ยังไม่ใช่ Document — slipreader เป็นคนประกอบ
// ค่า Level ใช้ค่าคงที่จาก package kind (high / low / missing)
type Result struct {
	Type              string
	AmountSatang      amount.Satang
	AmountLevel       string
	Counterparty      string
	CounterpartyLevel string
	OccurredAt        time.Time
	OccurredAtLevel   string
	Ref               string
}

// Extractor แกะฟิลด์จากบรรทัดข้อความ (ที่ผ่าน textnorm.CleanLines แล้ว) ของ issuer หนึ่ง ๆ
type Extractor interface {
	Extract(lines []string) Result
}

var extractors = map[string]Extractor{
	kind.KBank: kbank{},
}

// For คืนตัวแกะฟิลด์ของ issuer นั้น หรือตัวสำรองทั่วไปถ้าไม่รู้จัก
func For(issuer string) Extractor {
	if e, ok := extractors[issuer]; ok {
		return e
	}
	return generic{}
}

// emptyResult คือจุดเริ่มของทุกตัวแกะ: ทุกฟิลด์ missing จนกว่าจะหาเจอ
func emptyResult(docType string) Result {
	return Result{
		Type:              docType,
		AmountLevel:       kind.Missing,
		CounterpartyLevel: kind.Missing,
		OccurredAtLevel:   kind.Missing,
	}
}

// ---- helper ร่วม ----

// จำนวนบรรทัดถัดจากคำนำหน้าที่ยอมมองหาค่า — สลิปทุกใบวางค่าไว้ติดกับคำนำหน้า
// ถ้ามองไกลกว่านี้จะไปเจอ "ค่าธรรมเนียม 0.00" ที่อยู่ถัดลงมาแทน
const lookahead = 2

// indexOf คืนดัชนีของบรรทัดแรกที่มีคำนั้น หรือ -1
func indexOf(lines []string, substr string) int {
	for i, line := range lines {
		if strings.Contains(line, substr) {
			return i
		}
	}
	return -1
}

// amountAfter หายอดเงินที่อยู่บรรทัดเดียวกับคำนำหน้า (ส่วนที่ตามหลังคำ) หรือในอีกไม่เกิน lookahead บรรทัด
func amountAfter(lines []string, keyword string) (amount.Satang, bool) {
	i := indexOf(lines, keyword)
	if i < 0 {
		return 0, false
	}
	_, after, _ := strings.Cut(lines[i], keyword)
	if v, ok := amount.Parse(after); ok {
		return v, true
	}
	for j := i + 1; j <= i+lookahead && j < len(lines); j++ {
		if v, ok := amount.Parse(lines[j]); ok {
			return v, true
		}
	}
	return 0, false
}

// firstDate หาวันเวลาแรกในเอกสาร ลองทีละบรรทัด และลองต่อบรรทัดถัดไปอีก 1-2 บรรทัด
// เพราะเป๋าตังบางใบพิมพ์ "1 ก.ย." / "2569" / "12:35 น." แยกกันสามบรรทัด
func firstDate(lines []string) (time.Time, bool) {
	for i := range lines {
		joined := lines[i]
		for extra := 0; extra <= 2; extra++ {
			if extra > 0 {
				if i+extra >= len(lines) {
					break
				}
				joined += " " + lines[i+extra]
			}
			if t, ok := thaidate.Parse(joined); ok {
				return t, true
			}
		}
	}
	return time.Time{}, false
}

// รูปแบบเลขบัญชี/บัตร/PromptPay ที่ถูกปิดบังบางส่วน เช่น xxx-x-x1234-x, 4050-16XX-XXXX-9876, 081-xxx-2222
var accountPattern = regexp.MustCompile(`(?i)^[x\d]{1,4}(-[x\d]{1,8})+$`)

func looksLikeAccount(line string) bool {
	return accountPattern.MatchString(line) && strings.ContainsAny(line, "0123456789")
}

// nextTextLine คืนบรรทัดถัดจาก from ที่ "เป็นข้อความ" — มีตัวอักษรอย่างน้อย 3 ตัว และไม่ใช่เลขบัญชี
func nextTextLine(lines []string, from int) (string, bool) {
	for j := from + 1; j < len(lines); j++ {
		if textnorm.LetterCount(lines[j]) >= 3 && !looksLikeAccount(lines[j]) {
			return lines[j], true
		}
	}
	return "", false
}

// valueAfterLabel คืนข้อความหลังคำนำหน้าในบรรทัดเดียวกัน (ตัด ":" และช่องว่างออก)
// ถ้าบรรทัดนั้นมีแค่คำนำหน้า ให้คืนบรรทัดถัดไปแทน
func valueAfterLabel(lines []string, label string) (string, bool) {
	i := indexOf(lines, label)
	if i < 0 {
		return "", false
	}
	_, after, _ := strings.Cut(lines[i], label)
	after = strings.TrimSpace(strings.TrimLeft(strings.TrimSpace(after), ":"))
	if after != "" {
		return after, true
	}
	if i+1 < len(lines) {
		return lines[i+1], true
	}
	return "", false
}
