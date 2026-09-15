// Package textnorm ทำความสะอาดและเทียบข้อความที่ได้จาก OCR
//
// OCR ภาษาไทยมีอาการประจำสองอย่าง: เว้นวรรคผิดที่ และวรรณยุกต์เพี้ยนไปตัวเดียว
// ฟังก์ชันในนี้มีไว้ให้ส่วนอื่นเทียบข้อความได้โดยไม่ต้องสนอาการเหล่านั้น
package textnorm

import (
	"strings"
	"unicode"
)

// CleanLines ตัดช่องว่างหัวท้าย และทิ้งบรรทัดที่ OCR อ่านทะลุไอคอนหรือ QR ออกมาเป็นขยะ
//
// ถือว่าเป็นขยะเมื่อ: ไม่มีตัวอักษรหรือตัวเลขเลย (")", "+")
// หรือทั้งบรรทัดเป็นตัวอักษร/ตัวเลขตัวเดียวโดด ๆ ("0", "V")
// แต่ "K+" ต้องรอด เพราะเป็นคำที่ใช้บอกว่าสลิปมาจาก K PLUS
func CleanLines(raw []string) []string {
	cleaned := make([]string, 0, len(raw))
	for _, line := range raw {
		line = strings.TrimSpace(line)
		if isJunk(line) {
			continue
		}
		cleaned = append(cleaned, line)
	}
	return cleaned
}

func isJunk(line string) bool {
	alnum := 0
	for _, r := range line {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			alnum++
		}
	}
	if alnum == 0 {
		return true
	}
	return alnum == 1 && len([]rune(line)) == 1
}

// LetterCount นับตัวอักษร (ไม่นับตัวเลขและสัญลักษณ์) ใช้ตัดสินว่าบรรทัดนี้ "เป็นข้อความ" พอจะเป็นชื่อได้ไหม
func LetterCount(s string) int {
	count := 0
	for _, r := range s {
		if unicode.IsLetter(r) {
			count++
		}
	}
	return count
}

// Squash ตัดช่องว่างทุกชนิดออกทั้งหมด
func Squash(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, s)
}

// StripToneMarks ตัดวรรณยุกต์ ( ่ ้ ๊ ๋ ) และไม้ไต่คู้ ( ็ ) ออก
// เพราะเป็นตัวที่ OCR อ่านเพี้ยนบ่อยที่สุดในชื่อร้าน
func StripToneMarks(s string) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case '่', '้', '๊', '๋', '็':
			return -1
		}
		return r
	}, s)
}

// Similar เทียบข้อความสองชิ้นแบบหยาบ: ไม่สนช่องว่างและวรรณยุกต์
func Similar(a, b string) bool {
	return Squash(StripToneMarks(a)) == Squash(StripToneMarks(b))
}
