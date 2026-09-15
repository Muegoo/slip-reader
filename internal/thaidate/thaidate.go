// Package thaidate อ่านวันเวลาแบบที่สลิปไทยพิมพ์: เดือนย่อภาษาไทย ปีพุทธศักราช 2 หรือ 4 หลัก
package thaidate

import (
	"regexp"
	"strconv"
	"time"
)

// Bangkok คือเขตเวลาของทุกสลิปในระบบ
//
// ถ้าเครื่องไม่มี tzdata (container เล็ก ๆ บางตัว) ให้ใช้ +07:00 ตายตัวแทน — ไทยไม่มี DST จึงเท่ากันเสมอ
var Bangkok = loadBangkok()

func loadBangkok() *time.Location {
	loc, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		return time.FixedZone("ICT", 7*3600)
	}
	return loc
}

var thaiMonths = map[string]time.Month{
	"ม.ค.": time.January, "ก.พ.": time.February, "มี.ค.": time.March,
	"เม.ย.": time.April, "พ.ค.": time.May, "มิ.ย.": time.June,
	"ก.ค.": time.July, "ส.ค.": time.August, "ก.ย.": time.September,
	"ต.ค.": time.October, "พ.ย.": time.November, "ธ.ค.": time.December,
}

const yearPattern = `25\d{2}|20\d{2}|\d{2}`

var (
	// "29 พ.ค. 69" / "28ก.ค. 69" / "7 ก.ย.69" — ช่องว่างหายได้ทุกจุดเพราะ OCR
	// ปีรับ 25xx / 20xx / หรือ 2 หลัก ตามลำดับ — ห้ามใช้ \d{2,4} เพราะ "6907:35" (ปีติดเวลา) จะถูกกินเป็นปี 6907
	thaiDatePattern = regexp.MustCompile(`(\d{1,2})\s*(ม\.ค\.|ก\.พ\.|มี\.ค\.|เม\.ย\.|พ\.ค\.|มิ\.ย\.|ก\.ค\.|ส\.ค\.|ก\.ย\.|ต\.ค\.|พ\.ย\.|ธ\.ค\.)\s*(` + yearPattern + `)`)
	// "14/09/69" แบบใบเสร็จ 7-Eleven
	numericDatePattern = regexp.MustCompile(`(\d{1,2})/(\d{1,2})/(` + yearPattern + `)`)
	timePattern        = regexp.MustCompile(`(\d{1,2}):(\d{2})`)
)

// Parse อ่านวันและเวลาจากข้อความบรรทัดเดียว คืน false ถ้าไม่มีวันหรือไม่มีเวลา
//
// ต้องมีทั้งคู่ เพราะสลิปทุกใบมีเวลา — ถ้าวันกับเวลาอยู่คนละบรรทัด (เป๋าตังบางใบ)
// ผู้เรียกต้องต่อสองบรรทัดเข้าด้วยกันก่อน
func Parse(s string) (time.Time, bool) {
	day, month, year, dateEnd, ok := findDate(s)
	if !ok {
		return time.Time{}, false
	}
	timeMatch := timePattern.FindStringSubmatchIndex(s[dateEnd:])
	if timeMatch == nil {
		return time.Time{}, false
	}
	rest := s[dateEnd:]
	hour, _ := strconv.Atoi(rest[timeMatch[2]:timeMatch[3]])
	minute, _ := strconv.Atoi(rest[timeMatch[4]:timeMatch[5]])
	if hour > 23 || minute > 59 {
		return time.Time{}, false
	}
	return time.Date(year, month, day, hour, minute, 0, 0, Bangkok), true
}

// findDate คืน วัน/เดือน/ปี ค.ศ. และตำแหน่ง byte ที่วันจบ (เพื่อหาเวลาต่อจากตรงนั้น)
func findDate(s string) (day int, month time.Month, year int, end int, ok bool) {
	if m := thaiDatePattern.FindStringSubmatchIndex(s); m != nil {
		day, _ = strconv.Atoi(s[m[2]:m[3]])
		month = thaiMonths[s[m[4]:m[5]]]
		rawYear, _ := strconv.Atoi(s[m[6]:m[7]])
		return day, month, buddhistToGregorian(rawYear), m[1], validDay(day)
	}
	if m := numericDatePattern.FindStringSubmatchIndex(s); m != nil {
		day, _ = strconv.Atoi(s[m[2]:m[3]])
		monthNumber, _ := strconv.Atoi(s[m[4]:m[5]])
		rawYear, _ := strconv.Atoi(s[m[6]:m[7]])
		if monthNumber < 1 || monthNumber > 12 {
			return 0, 0, 0, 0, false
		}
		return day, time.Month(monthNumber), buddhistToGregorian(rawYear), m[1], validDay(day)
	}
	return 0, 0, 0, 0, false
}

func validDay(day int) bool {
	return day >= 1 && day <= 31
}

// buddhistToGregorian แปลงปีที่สลิปพิมพ์เป็นปี ค.ศ.
//
// 2 หลัก ("69") = พ.ศ. 25xx → 2569 → 2026 · 4 หลักตั้งแต่ 2400 = พ.ศ. → ลบ 543 · อื่น ๆ ถือว่าเป็น ค.ศ. อยู่แล้ว
func buddhistToGregorian(year int) int {
	if year < 100 {
		return year + 2500 - 543
	}
	if year >= 2400 {
		return year - 543
	}
	return year
}
