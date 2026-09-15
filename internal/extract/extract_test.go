package extract

import (
	"testing"

	"github.com/Muegoo/slip-reader/internal/amount"
)

func TestAmountAfter(t *testing.T) {
	cases := []struct {
		name    string
		lines   []string
		keyword string
		want    amount.Satang
		ok      bool
	}{
		{"บรรทัดเดียวกัน", []string{"จำนวน: 119.00 บาท"}, "จำนวน", 11900, true},
		{"บรรทัดถัดไป", []string{"จำนวน:", "199.00 บาท", "ค่าธรรมเนียม:", "0.00 บาท"}, "จำนวน", 19900, true},
		{"ข้ามบรรทัดที่ไม่ใช่เงินแล้วเจอในบรรทัดที่สอง", []string{"จำนวนเงินที่ชำระ", "บาท", "20"}, "จำนวนเงินที่ชำระ", 2000, true},
		{"ไกลเกิน 2 บรรทัดไม่นับ", []string{"จำนวน:", "a", "b", "50.00"}, "จำนวน", 0, false},
		{"ไม่มี keyword", []string{"199.00 บาท"}, "จำนวน", 0, false},
		{"เลขติดลบถูกข้าม", []string{"สิทธิ", "-30 บาท", "20 บาท"}, "สิทธิ", 2000, true},
	}
	for _, c := range cases {
		got, ok := amountAfter(c.lines, c.keyword)
		if ok != c.ok || got != c.want {
			t.Errorf("%s: amountAfter = (%d, %v) ต้องการ (%d, %v)", c.name, got, ok, c.want, c.ok)
		}
	}
}

func TestFirstDateJoinsSplitLines(t *testing.T) {
	// เป๋าตังบางใบ วัน/ปี/เวลา แยกกัน 3 บรรทัด
	got, ok := firstDate([]string{"ทำรายการสำเร็จ", "1 ก.ย.", "2569", "12:35 น.", "สมชาย"})
	if !ok || got.Hour() != 12 || got.Minute() != 35 || got.Day() != 1 {
		t.Errorf("firstDate = %v, %v", got, ok)
	}
	if _, ok := firstDate([]string{"จำนวน:", "119.00 บาท"}); ok {
		t.Error("firstDate ต้องไม่เจอวันในบรรทัดที่ไม่มีวัน")
	}
}

func TestNextTextLineSkipsAccountsAndJunk(t *testing.T) {
	lines := []string{"ธ.กสิกรไทย", "xxX-x-x1234-x", "12", "นายสมหญิง รักดี", "ธ.กสิกรไทย"}
	got, ok := nextTextLine(lines, 0)
	if !ok || got != "นายสมหญิง รักดี" {
		t.Errorf("nextTextLine = %q, %v", got, ok)
	}
	if _, ok := nextTextLine(lines, len(lines)-1); ok {
		t.Error("nextTextLine หลังบรรทัดสุดท้ายต้องไม่เจอ")
	}
}

func TestLooksLikeAccount(t *testing.T) {
	yes := []string{"xxX-x-x1234-x", "XXX-X-XX123-4", "006-xXXXXXXX-2468", "081-xxx-2222", "4050-16XX-XXXX-9876", "x-1111", "012-3-xxx456"}
	no := []string{"Payatai Plaza", "นายสมหญิง รักดี", "199.00 บาท", "K+", "2609011851483324039"}
	for _, s := range yes {
		if !looksLikeAccount(s) {
			t.Errorf("looksLikeAccount(%q) ต้องเป็น true", s)
		}
	}
	for _, s := range no {
		if looksLikeAccount(s) {
			t.Errorf("looksLikeAccount(%q) ต้องเป็น false", s)
		}
	}
}
