package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"time"

	"github.com/Muegoo/slip-reader/internal/amount"
	"github.com/Muegoo/slip-reader/internal/kind"
)

// Label คือหนึ่งแถวของ labels.csv — ground truth ที่คนอ่านจากรูปจริงแล้วกรอกไว้
type Label struct {
	File         string
	IsSlip       bool
	Issuer       string // ค่าคงที่จาก package kind
	AmountSatang amount.Satang
	Counterparty string
	Date         time.Time // เฉพาะวัน ไม่มีเวลา — labels.csv ไม่ได้บันทึกเวลาไว้
}

// รูปแบบ CSV ตรงกับที่ spike สร้างไว้ (file,is_slip,source,amount,counterparty,date,note)
// เพื่อให้คัดลอกไฟล์จากคลังสลิปเดิมมาใช้ได้ทันทีโดยไม่ต้องแปลง
var sourceToIssuer = map[string]string{
	"K PLUS":          kind.KBank,
	"เป๋าตัง":         kind.Paotang,
	"ttb":             kind.TTB,
	"7-Eleven (7App)": kind.SevenEleven,
	"Bangkok Bank":    kind.BBL,
	"Dime":            kind.Dime,
}

func readLabels(r io.Reader) ([]Label, error) {
	reader := csv.NewReader(r)
	// คอลัมน์ note เป็นข้อความอิสระที่คนพิมพ์ อาจมีเครื่องหมายคำพูดหลุดมา — ไม่ควรทำให้ทั้งไฟล์อ่านไม่ได้
	reader.LazyQuotes = true
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("อ่าน CSV: %w", err)
	}
	if len(rows) < 2 {
		return nil, fmt.Errorf("labels.csv ว่าง")
	}

	labels := make([]Label, 0, len(rows)-1)
	for lineNo, row := range rows[1:] {
		label, err := parseLabelRow(row)
		if err != nil {
			return nil, fmt.Errorf("labels.csv บรรทัด %d: %w", lineNo+2, err)
		}
		labels = append(labels, label)
	}
	return labels, nil
}

func parseLabelRow(row []string) (Label, error) {
	if len(row) < 6 {
		return Label{}, fmt.Errorf("มี %d คอลัมน์ ต้องการอย่างน้อย 6", len(row))
	}
	label := Label{File: row[0], IsSlip: row[1] == "yes"}
	if !label.IsSlip {
		return label, nil
	}

	issuer, ok := sourceToIssuer[row[2]]
	if !ok {
		return Label{}, fmt.Errorf("ไม่รู้จัก source %q — เพิ่มใน sourceToIssuer ก่อน", row[2])
	}
	label.Issuer = issuer

	satang, ok := amount.Parse(row[3])
	if !ok {
		return Label{}, fmt.Errorf("ยอด %q อ่านไม่ได้", row[3])
	}
	label.AmountSatang = satang
	label.Counterparty = row[4]

	date, err := time.Parse("2006-01-02", row[5])
	if err != nil {
		return Label{}, fmt.Errorf("วันที่ %q: %w", row[5], err)
	}
	label.Date = date
	return label, nil
}
