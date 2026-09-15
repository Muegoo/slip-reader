# slip-reader

> Status: in active development — ตัวแกะฟิลด์ผ่าน accuracy harness บนสลิปจริง 65 ใบแล้ว
> ส่วน OCR container (`ocr-service/`) ยังรอยืนยันด้วย `docker compose up` จริง

Go library ที่รับรูป**สลิปโอนเงิน**หรือ**ใบเสร็จ**ของไทย แล้วคืน ยอดเงิน · คู่ค้า · เวลา · เลขอ้างอิง
ในรูปแบบเดียว ไม่ว่าจะมาจากธนาคารไหนหรือร้านอะไร พร้อมบอกความมั่นใจแยกทีละฟิลด์

ทำไมต้องมี: คนที่ใช้จ่ายผ่านการโอนเป็นหลักมีสลิปเต็มมือถือแต่ไม่มีใครมานั่งพิมพ์ยอดลงแอปบันทึกรายจ่าย
ไลบรารีนี้คือชิ้นส่วนที่ทำให้ "แชร์สลิปเข้าแชท → บันทึกให้เอง" เป็นไปได้ (ใช้ใน [banchee](../banchee))

## ใช้งาน

```go
import (
    slipreader "github.com/Muegoo/slip-reader"
    "github.com/Muegoo/slip-reader/ocr"
)

reader := slipreader.New(ocr.NewPaddle("http://localhost:8866"))
doc, err := reader.Read(ctx, imageBytes)
// doc.AmountSatang = 11900  (หน่วยสตางค์เสมอ — 119.00 บาท)
// doc.Counterparty = "บ.เซ็นทรัล เรสตอรองส์ กรุ๊ป จก."
// doc.Confidence   = {Amount: high, Counterparty: high, OccurredAt: high}
```

`Read` **ไม่คืน error เพราะอ่านไม่ออก** — รูปที่ไม่ใช่สลิปได้ `Document` ที่ทุกฟิลด์เป็น `missing`
error สงวนไว้ให้ OCR ล่มหรือ ctx หมดเวลา ผู้เรียกจึงแยก "ระบบพัง" กับ "อ่านไม่ออก" ได้เสมอ

### ลองจากบรรทัดคำสั่ง

```sh
docker compose up -d --build            # OCR engine (PaddleOCR) ที่ localhost:8866 — build ครั้งแรกโหลดโมเดล ~10 นาที
go run ./cmd/slipread path/to/slip.jpg  # พิมพ์ Document เป็น JSON
```

## ความแม่น

วัดด้วย `go run ./cmd/accuracy` บนสลิปจริง 63 ใบ + รูปที่ไม่ใช่สลิป 2 ใบ (2026-09-15, PaddleOCR mobile):

| issuer | ใบ | issuer ถูก | ยอด | คู่ค้า | วันที่ |
|---|---|---|---|---|---|
| K PLUS | 26 | 26/26 | 26/26 | 25/26 | 26/26 |
| เป๋าตัง | 13 | 13/13 | 13/13 | 13/13 | 13/13 |
| ttb | 10 | 10/10 | 10/10 | 10/10 | 10/10 |
| 7-Eleven | 7 | 7/7 | 7/7 | 7/7 | 7/7 |
| Bangkok Bank | 4 | 4/4 | 4/4 | 4/4 | 4/4 |
| Dime | 3 | 3/3 | 3/3 | 3/3 | 2/3 |
| **รวม** | 63 | **63/63** | **63/63** | **62/63** | **62/63** |

รูปที่ไม่ใช่สลิป: 0/2 ที่ระบบเดายอดออกมา · 2 ใบที่พลาดคือ OCR อ่านเพี้ยนจริง (`1lก.ค.` แทน `11 ก.ค.` → ระบบตอบ
`missing` ไม่เดา, `สาขาา` เกินสระหนึ่งตัว) · QR ถอดได้ครบทุกใบที่มี (33/33)

ตัวเลขนี้มาจากสลิปของคนคนเดียว — ธนาคารและร้านที่ไม่อยู่ในตารางยังไม่เคยเจอ ถ้ามีสลิปที่อ่านผิดช่วยเปิด issue
(**ปิดบังชื่อและเลขบัญชีก่อน**)

## ทำงานอย่างไร

```
รูป ──► ocr.Engine (PaddleOCR หลัง HTTP) ──► textnorm.CleanLines ──► issuer.Detect ──► extract.For(issuer)
 │                                                                                          │
 └──► qr.Decode ──────────────────────────────── เลขอ้างอิงจาก QR แทนที่ของ OCR ถ้ามี ──────┴──► Document
```

- **OCR อยู่หลัง interface** (`ocr.Engine`) — เทสต์ทั้งหมดใช้ `ocr.Fake` และข้อความ OCR ที่บันทึกไว้ ไม่ต้องรัน container
- **ตัวแกะฟิลด์หนึ่งไฟล์ต่อ issuer** (`internal/extract/kbank.go` …) แต่ละไฟล์บอกว่าฟิลด์อยู่ตรงไหน ~60 บรรทัด
- **ยอดเงินเลือกจากคำนำหน้า ไม่ใช่ตัวเลขใหญ่สุด** — ใบเดียวมีหลายยอด (ราคา / ส่วนลด / จ่ายจริง / ค่าธรรมเนียม 0.00)
- **เงินเป็น `int64` สตางค์เสมอ** ไม่มี float ในเส้นทางของเงิน
- issuer ที่ไม่รู้จักตกไปที่ตัวสำรองทั่วไป: อ่านได้เท่าที่อ่านได้ (level `low`) และไม่เดาชื่อคู่ค้า

## เพิ่ม issuer ใหม่

1. เก็บข้อความ OCR ของสลิปนั้นไว้ที่ `testdata/ocr/<issuer>/<case>.txt` **ปิดบังข้อมูลส่วนบุคคลตามกติกาใน
   [`testdata/ocr/README.md`](testdata/ocr/README.md)** และเขียน `<case>.json` ที่คาดไว้จากรูปจริง
2. เพิ่มกติกาเดา issuer ใน `internal/issuer/issuer.go` (คำที่บอก*ผู้ออกสลิป* ไม่ใช่ธนาคารผู้รับ)
3. สร้าง `internal/extract/<issuer>.go` แล้วลงทะเบียนใน `extractors` — `go test ./...` จะรัน golden test ให้เอง

## ทดสอบ

```sh
go test ./...                                      # ไม่ต้องมี container
go run ./cmd/accuracy -slips DIR -labels DIR/labels.csv   # ต้องมี container หรือ cache ใน testdata/ocr-cache
```

`testdata/slips-raw/` และ `testdata/ocr-cache/` อยู่ใน `.gitignore` — **สลิปจริงห้ามเข้า repo นี้เด็ดขาด**
เพราะมีชื่อและเลขบัญชีของคนอื่น สิ่งที่ commit ได้มีแค่ข้อความ OCR ที่ปิดบังแล้วใน `testdata/ocr/`

## สิ่งที่ยังไม่ทำ

- ปรับรูปก่อน OCR (แก้เอียง/คอนทราสต์) — ชุดข้อมูลตอนนี้เป็น screenshot ล้วน ยังไม่มีสลิปกระดาษถ่ายกล้อง
- issuer อื่นนอก 6 รายข้างบน (SCB, KTB, TrueMoney, ShopeePay …) — รอมีสลิปจริง
- ตัวสำรองทั่วไปยังไม่พยายามอ่านชื่อคู่ค้าเลย

## License

MIT
