// Package slipreader อ่านรูปสลิปโอนเงินและใบเสร็จของไทย แล้วคืนข้อมูลรูปแบบเดียว
// ไม่ว่าเอกสารจะมาจากธนาคารไหนหรือร้านอะไร
//
// ไลบรารีนี้ไม่ยืนยันกับธนาคาร ไม่จัดหมวดหมู่ ไม่เก็บอะไรลงฐานข้อมูล และไม่รู้จักผู้ใช้
// รับรูป คืนข้อมูล จบ — ความไม่รู้อะไรเลยนี่แหละที่ทำให้มันเทสต์ง่ายและเอาไปใช้ที่อื่นได้
package slipreader

import (
	"context"
	"time"
)

// DocType บอกว่าเอกสารเป็นสลิปโอนเงินหรือใบเสร็จร้านค้า
type DocType string

const (
	DocTypeTransfer DocType = "transfer"
	DocTypeReceipt  DocType = "receipt"
	DocTypeUnknown  DocType = "unknown"
)

// Issuer คือแอปหรือร้านที่ออกเอกสาร เรียงตามจำนวนที่พบจริงในชุดข้อมูลของเจ้าของโปรเจค
type Issuer string

const (
	IssuerKBank       Issuer = "kbank"   // K PLUS
	IssuerPaotang     Issuer = "paotang" // เป๋าตัง
	IssuerTTB         Issuer = "ttb"
	IssuerSevenEleven Issuer = "7eleven" // ใบเสร็จใน 7App
	IssuerBBL         Issuer = "bbl"     // Bangkok Bank
	IssuerDime        Issuer = "dime"
	IssuerUnknown     Issuer = "unknown"
)

// Level คือความมั่นใจของฟิลด์หนึ่ง ๆ
//
//   - high    อ่านได้จากกติกาของ issuer ที่รู้จัก ตรงคำนำหน้าที่คาดไว้
//   - low     ได้มาจากตัวสำรองทั่วไป หรือกติกาของ issuer ต้องเดาบางส่วน — ควรให้ผู้ใช้ยืนยัน
//   - missing หาไม่เจอเลย ค่าในฟิลด์นั้นไม่มีความหมาย
type Level string

const (
	LevelHigh    Level = "high"
	LevelLow     Level = "low"
	LevelMissing Level = "missing"
)

// Confidence บอกความมั่นใจแยกทีละฟิลด์ เพราะ UX ต้องรู้ว่า "ยอดชัดแต่ชื่อร้านอ่านไม่ออก"
// ไม่ใช่แค่ตัวเลขรวมตัวเดียว
type Confidence struct {
	Amount       Level `json:"amount"`
	Counterparty Level `json:"counterparty"`
	OccurredAt   Level `json:"occurred_at"`
}

// Document คือผลการอ่านหนึ่งเอกสาร
type Document struct {
	Type         DocType    `json:"type"`
	Issuer       Issuer     `json:"issuer"`
	AmountSatang int64      `json:"amount_satang"` // หน่วยสตางค์เสมอ 0 ถ้า Confidence.Amount เป็น missing
	OccurredAt   time.Time  `json:"occurred_at"`   // zero ถ้า missing
	Counterparty string     `json:"counterparty"`  // สลิปโอน = ผู้รับ / ใบเสร็จ = ร้าน + สาขา
	Ref          string     `json:"ref"`           // เลขอ้างอิง ถ้ามี (ว่างได้)
	Confidence   Confidence `json:"confidence"`
	RawLines     []string   `json:"raw_lines"` // ข้อความดิบจาก OCR ทีละบรรทัด — ให้แอปเก็บไว้รันใหม่ทีหลังได้
}

// Reader อ่านรูปหนึ่งรูป
//
// Read ไม่คืน error เพราะ "อ่านไม่ออก" — กรณีนั้นคืน Document ที่ Confidence ทุกฟิลด์เป็น missing
// error สงวนไว้ให้ปัญหาเชิงระบบ เช่น OCR engine ล่ม รูปเปิดไม่ได้ หรือ ctx หมดเวลา
type Reader interface {
	Read(ctx context.Context, image []byte) (*Document, error)
}
