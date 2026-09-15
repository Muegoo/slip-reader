// Package kind เก็บค่าคงที่ที่ทั้ง package ราก (slipreader) และ package ย่อยใน internal/ ต้องใช้ร่วมกัน
//
// ทำไมต้องแยกมาอยู่ตรงนี้: slipreader import internal/issuer และ internal/extract
// ถ้าสองตัวนั้น import slipreader กลับเพื่อเอา type Issuer/Level จะเกิด import cycle
// package นี้จึงเป็นใบล่างสุดที่ไม่ import อะไรเลย ส่วน slipreader ห่อค่าเหล่านี้เป็น type สาธารณะอีกชั้น
package kind

const (
	DocTransfer = "transfer"
	DocReceipt  = "receipt"
	DocUnknown  = "unknown"
)

const (
	KBank       = "kbank"
	Paotang     = "paotang"
	TTB         = "ttb"
	SevenEleven = "7eleven"
	BBL         = "bbl"
	Dime        = "dime"
	Unknown     = "unknown"
)

const (
	High    = "high"
	Low     = "low"
	Missing = "missing"
)
