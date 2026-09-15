// Package amount แปลงข้อความยอดเงินที่ OCR อ่านได้ให้เป็นสตางค์
//
// เก็บเป็นจำนวนเต็มหน่วยสตางค์เสมอ ไม่ผ่าน float แม้แต่ตอน parse — เหตุผลเดียวกับ
// package money ของ banchee: ทศนิยมฐานสองแทน 0.01 ไม่ตรง พอสะสมหลายรายการยอดจะเพี้ยน
package amount

import (
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// Satang คือจำนวนเงินหน่วยสตางค์ 100 สตางค์เท่ากับ 1 บาท
type Satang int64

// Pattern จับกลุ่มตัวเลขที่อาจเป็นยอดเงิน (ตัวเลขล้วน คั่นด้วยจุดหรือจุลภาคได้)
// ใช้เช็คหยาบ ๆ ว่า "บรรทัดนี้มีตัวเลขหน้าตาเหมือนเงินไหม" — การตัดสินจริงอยู่ที่ Parse
var Pattern = regexp.MustCompile(`\d(?:[\d.,]*\d)?`)

// เลขที่ไม่มีตัวคั่นและยาวเกินนี้คือเลขอ้างอิง ไม่ใช่เงิน (สลิปยอดสูงสุดที่คิดว่าจะเจอคือหลักล้าน)
const maxPlainDigits = 7

// Parse หาตัวเลขเงินตัวแรกในข้อความ คืน false ถ้าไม่มีตัวไหนผ่านเกณฑ์
//
// ปฏิเสธ: เลขติดลบ (ส่วนลด), เลขที่ติดกับตัวอักษรหรือขีดกลางทางซ้าย (เลขบัญชี xxx-x-x1234-x),
// เลขที่ตามด้วยขีดกลาง (081-xxx-2222), เลขนำหน้าด้วยศูนย์ (016244…), และเลขยาวไม่มีตัวคั่น (TID#2026…)
func Parse(s string) (Satang, bool) {
	for _, loc := range Pattern.FindAllStringIndex(s, -1) {
		token := s[loc[0]:loc[1]]
		before := runeBefore(s, loc[0])
		after := runeAfter(s, loc[1])

		if before == '-' || after == '-' || unicode.IsLetter(before) {
			continue
		}
		if looksLikeReference(token) {
			continue
		}
		whole, frac := splitDecimal(token)
		if len(whole) > 1 && whole[0] == '0' {
			continue
		}
		return toSatang(whole, frac)
	}
	return 0, false
}

func runeBefore(s string, byteIndex int) rune {
	if byteIndex == 0 {
		return 0
	}
	prefix := []rune(s[:byteIndex])
	return prefix[len(prefix)-1]
}

func runeAfter(s string, byteIndex int) rune {
	if byteIndex >= len(s) {
		return 0
	}
	rest := []rune(s[byteIndex:])
	return rest[0]
}

func looksLikeReference(token string) bool {
	hasSeparator := strings.ContainsAny(token, ".,")
	return !hasSeparator && len(token) > maxPlainDigits
}

// splitDecimal แยกส่วนบาทกับส่วนสตางค์ออกจากกัน โดยไม่สนว่าตัวคั่นเป็นจุดหรือจุลภาค
//
// กติกา: ชิ้นสุดท้ายเป็นทศนิยมก็ต่อเมื่อยาว 1-2 หลัก ("31.700.00" → 31700 กับ 00)
// ถ้ายาว 3 หลักถือว่าเป็นหลักพันทั้งหมด ("1.234" → 1234) เพราะไม่มีสลิปไหนแสดงสตางค์ 3 หลัก
func splitDecimal(token string) (whole, frac string) {
	parts := strings.FieldsFunc(token, func(r rune) bool { return r == '.' || r == ',' })
	if len(parts) == 1 {
		return parts[0], ""
	}
	last := parts[len(parts)-1]
	if len(last) <= 2 {
		return strings.Join(parts[:len(parts)-1], ""), last
	}
	return strings.Join(parts, ""), ""
}

func toSatang(whole, frac string) (Satang, bool) {
	baht, err := strconv.ParseInt(whole, 10, 64)
	if err != nil {
		return 0, false
	}
	// "6" หมายถึง 60 สตางค์ ไม่ใช่ 6 — เติมศูนย์ทางขวาให้ครบสองหลักก่อน
	for len(frac) < 2 {
		frac += "0"
	}
	satang, err := strconv.ParseInt(frac, 10, 64)
	if err != nil {
		return 0, false
	}
	return Satang(baht*100 + satang), true
}
