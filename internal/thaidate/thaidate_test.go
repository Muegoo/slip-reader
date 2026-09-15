package thaidate

import (
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	// ทุกรูปแบบมาจากผล OCR จริงตอน spike
	cases := []struct {
		in   string
		want string // RFC3339 ใน ICT
	}{
		{"29 พ.ค. 69, 09:28", "2026-05-29T09:28:00+07:00"},       // BBL
		{"29 พ.ค. 69  09:29 น.", "2026-05-29T09:29:00+07:00"},    // K PLUS ช่องว่างสองตัว
		{"1 ก.ย. 2569 12:35 น.", "2026-09-01T12:35:00+07:00"},    // เป๋าตัง ปีเต็ม
		{"16 พ.ค. 2569 - 11:45 น.", "2026-05-16T11:45:00+07:00"}, // Dime
		{"14/09/69 | 19:47", "2026-09-14T19:47:00+07:00"},        // 7-Eleven
		{"14/09/69 119:47", "2026-09-14T19:47:00+07:00"},         // 7-Eleven ที่ Paddle อ่าน | เป็น 1
		{"28ก.ค. 69, 18:31 น.", "2026-07-28T18:31:00+07:00"},     // Paddle ตัดช่องว่างหาย
		{"7 ก.ย.69 19:13 น.", "2026-09-07T19:13:00+07:00"},       // ช่องว่างหน้าปีหาย
		{"31 พ.ค. 69  19:24 น.", "2026-05-31T19:24:00+07:00"},
		{"2 ก.ย. 2569 12:21 น.", "2026-09-02T12:21:00+07:00"},           // สองบรรทัดที่ผู้เรียก join แล้ว
		{"วันที่ 16 พ.ค. 2569 - 11:45 น.", "2026-05-16T11:45:00+07:00"}, // มีคำนำหน้า
	}
	for _, c := range cases {
		got, ok := Parse(c.in)
		if !ok {
			t.Errorf("Parse(%q) ไม่ผ่าน", c.in)
			continue
		}
		if got.Format(time.RFC3339) != c.want {
			t.Errorf("Parse(%q) = %s ต้องการ %s", c.in, got.Format(time.RFC3339), c.want)
		}
	}
}

func TestParseRejectsNonDates(t *testing.T) {
	for _, in := range []string{
		"จำนวน: 119.00 บาท",
		"016244121632DQR09258",
		"2 ก.ย. 2569", // มีวันแต่ไม่มีเวลา — ผู้เรียกต้อง join บรรทัดเวลาให้ก่อน
		"12:21 น.",
		"",
	} {
		if got, ok := Parse(in); ok {
			t.Errorf("Parse(%q) = %v ต้องไม่ผ่าน", in, got)
		}
	}
}

func TestBuddhistToGregorian(t *testing.T) {
	cases := map[int]int{69: 2026, 2569: 2026, 2026: 2026, 99: 2056, 0: 1957}
	for in, want := range cases {
		if got := buddhistToGregorian(in); got != want {
			t.Errorf("buddhistToGregorian(%d) = %d ต้องการ %d", in, got, want)
		}
	}
}
