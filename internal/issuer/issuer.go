// Package issuer เดาว่าเอกสารมาจากแอปหรือร้านไหน โดยดูจากคำในข้อความ OCR
//
// ไม่ดูสี ไม่ดูตำแหน่งโลโก้ เพราะสลิปแอปเดียวกันมีหลายธีม (ttb สว่าง/มืด, K PLUS ลายน้ำตึก/การ์ตูน)
// แต่คำที่พิมพ์บนสลิปเหมือนกันทุกธีม
package issuer

import (
	"strings"

	"github.com/Muegoo/slip-reader/internal/kind"
)

type rule struct {
	issuer string
	match  func(lines []string) bool
}

// ลำดับสำคัญ: สลิปโอนเงินมีชื่อธนาคาร *ผู้รับ* อยู่ด้วย (K PLUS โอนไป Dime มีคำ "เกียรตินาคินภัทร")
// จึงต้องเช็คคำที่บอกตัว *ผู้ออกสลิป* ก่อนเสมอ และเช็คคำที่กำกวมที่สุด (ttb) ท้ายสุด
var rules = []rule{
	{kind.KBank, looksLikeKBank},
	{kind.Paotang, looksLikePaotang},
	{kind.SevenEleven, looksLikeSevenEleven},
	{kind.BBL, looksLikeBBL},
	{kind.Dime, looksLikeDime},
	{kind.TTB, looksLikeTTB},
}

// Detect คืนชื่อ issuer (ค่าคงที่ใน package kind) หรือ kind.Unknown ถ้าไม่เข้าเกณฑ์ไหนเลย
func Detect(lines []string) string {
	for _, r := range rules {
		if r.match(lines) {
			return r.issuer
		}
	}
	return kind.Unknown
}

func looksLikeKBank(lines []string) bool {
	return hasExactLine(lines, "K+") || hasSubstring(lines, "ธ.กสิกรไทย")
}

func looksLikePaotang(lines []string) bool {
	return hasSubstring(lines, "เป๋าตัง") || hasSubstring(lines, "เปาตัง") ||
		hasSubstring(lines, "G-Wallet ID") || hasSubstring(lines, "ไทยช่วยไทย")
}

func looksLikeSevenEleven(lines []string) bool {
	// OCR อ่าน n ท้ายคำเป็น ท บ่อย จึงเช็คแค่ส่วนหน้า
	return hasSubstring(lines, "7-Eleve") || hasSubstring(lines, "7Delivery") || hasSubstring(lines, "7App")
}

func looksLikeBBL(lines []string) bool {
	// ชื่อธนาคารกรุงเทพต้องอยู่ต้นสลิป (เป็นหัว) ถ้าอยู่ล่าง ๆ อาจเป็นแค่ธนาคารผู้รับของสลิปอื่น
	head := lines
	if len(head) > 3 {
		head = head[:3]
	}
	return hasSubstring(head, "Bangkok Bank") || hasSubstring(head, "ธนาคารกรุงเทพ")
}

func looksLikeDime(lines []string) bool {
	// "โดย ธ.เกียรตินาคินภัทร" มีคำว่า "โดย" = ผู้ออกสลิป ต่างจากชื่อธนาคารในช่องผู้รับของสลิปอื่น
	return hasSubstring(lines, "ime!") || hasSubstring(lines, "Dime") || hasSubstring(lines, "โดย ธ.เกียรตินาคินภัทร")
}

func looksLikeTTB(lines []string) bool {
	// โลโก้ ttb ถูกอ่านเป็นบรรทัดสั้น ๆ "ttb" หรือ "utb" — ต้องเป็นบรรทัดเดี่ยว ไม่ใช่ส่วนของคำอื่น
	for _, line := range lines {
		lower := strings.ToLower(line)
		if lower == "ttb" || lower == "utb" {
			return true
		}
	}
	return false
}

func hasExactLine(lines []string, want string) bool {
	for _, line := range lines {
		if line == want {
			return true
		}
	}
	return false
}

func hasSubstring(lines []string, want string) bool {
	for _, line := range lines {
		if strings.Contains(line, want) {
			return true
		}
	}
	return false
}
