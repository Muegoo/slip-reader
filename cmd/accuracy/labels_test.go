package main

import (
	"strings"
	"testing"
)

func TestReadLabels(t *testing.T) {
	csv := strings.TrimSpace(`
file,is_slip,source,amount,counterparty,date,note
IMG_2842.JPG,yes,Bangkok Bank,29311.00,นาย สมชาย ใจดี,2026-05-29,PromptPay
IMG_3289.JPG,yes,เป๋าตัง,20.00,โชคเทพขุนแผน,2026-09-01,"ไทยช่วยไทย: 50, -30, 20"
IMG_3298.JPG,no,-,,,,รูปแมว
`) + "\n"

	labels, err := readLabels(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("readLabels: %v", err)
	}
	if len(labels) != 3 {
		t.Fatalf("ได้ %d แถว ต้องการ 3", len(labels))
	}

	bbl := labels[0]
	if !bbl.IsSlip || bbl.Issuer != "bbl" || bbl.AmountSatang != 2931100 || bbl.Counterparty != "นาย สมชาย ใจดี" {
		t.Errorf("แถว BBL = %+v", bbl)
	}
	if got := bbl.Date.Format("2006-01-02"); got != "2026-05-29" {
		t.Errorf("วันที่ BBL = %s", got)
	}
	if labels[1].Issuer != "paotang" || labels[1].AmountSatang != 2000 {
		t.Errorf("แถวเป๋าตัง = %+v", labels[1])
	}
	cat := labels[2]
	if cat.IsSlip || cat.File != "IMG_3298.JPG" {
		t.Errorf("แถวรูปแมว = %+v", cat)
	}
}

func TestReadLabelsRejectsUnknownSource(t *testing.T) {
	csv := "file,is_slip,source,amount,counterparty,date,note\nx.jpg,yes,ธนาคารลึกลับ,1.00,a,2026-01-01,\n"
	if _, err := readLabels(strings.NewReader(csv)); err == nil || !strings.Contains(err.Error(), "ธนาคารลึกลับ") {
		t.Errorf("ต้อง error พร้อมบอกชื่อ source ที่ไม่รู้จัก ได้ %v", err)
	}
}
