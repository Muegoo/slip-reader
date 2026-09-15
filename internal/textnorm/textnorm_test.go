package textnorm

import (
	"slices"
	"testing"
)

func TestCleanLinesDropsJunk(t *testing.T) {
	// บรรทัดขยะทั้งหมดนี้มาจากผล OCR จริงตอน spike: OCR อ่านทะลุไอคอนและ QR ออกมาเป็นตัวเดี่ยว ๆ
	got := CleanLines([]string{"ชำระเงินสำเร็จ", "K+", "", "  ", "V", ")", "0", "119.00 บาท", "+", "  จำนวน:  "})
	want := []string{"ชำระเงินสำเร็จ", "K+", "119.00 บาท", "จำนวน:"}
	if !slices.Equal(got, want) {
		t.Errorf("CleanLines = %q ต้องการ %q", got, want)
	}
}

func TestSquashRemovesAllWhitespace(t *testing.T) {
	if got := Squash(" นาย สมชาย\tใจดี \n"); got != "นายสมชายใจดี" {
		t.Errorf("Squash = %q", got)
	}
}

func TestStripToneMarks(t *testing.T) {
	if got := StripToneMarks("มี่เสวี่ย โจ๊ก แอ็ด ต๋อม ป้า วงศ์ษา"); got != "มีเสวีย โจก แอด ตอม ปา วงศษา" {
		t.Errorf("StripToneMarks = %q", got)
	}
}

func TestSimilarIgnoresToneMarksAndSpaces(t *testing.T) {
	cases := []struct {
		a, b string
		want bool
	}{
		{"มี่เสวีย", "มี่เสวี่ย", true},                                  // spike: Paddle อ่านวรรณยุกต์หาย
		{"พี่ทีที่สเตชัน-สาขาจอมทอง", "พีทีที สเตชั่น-สาขาจอมทอง", true}, // สระ/วรรณยุกต์เพี้ยน + ช่องว่าง
		{"ข้าวไข่เจียวแม่แอ็ดพญาไท", "ข้าวไข่เจียวแม่แอ๊ดพญาไท", true},
		{"Payatai Plaza", "Payatai Plaza", true},
		{"บ.เซ็นทรัล เรสตอรองส์กรุ๊ปจก", "บ.เซ็นทรัล เรสตอรองส์ กรุ๊ป จก.", true}, // OCR ตัดจุดและช่องว่างหาย
		{"นาย สมชาย วงศดี", "นาย สมชาย วงศ์ดี", true},                             // การันต์หาย
		{"โจ๊กพญาไท", "ชูครีมมัช", false},
		{"", "", true},
	}
	for _, c := range cases {
		if got := Similar(c.a, c.b); got != c.want {
			t.Errorf("Similar(%q, %q) = %v ต้องการ %v", c.a, c.b, got, c.want)
		}
	}
}

func TestLetterCount(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{
		{"K+", 1},
		{"xxx-x-x1234-x", 6},
		{"นาย สมชาย", 8},
		{")", 0},
	}
	for _, c := range cases {
		if got := LetterCount(c.in); got != c.want {
			t.Errorf("LetterCount(%q) = %d ต้องการ %d", c.in, got, c.want)
		}
	}
}
