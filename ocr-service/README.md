# ocr-service

PaddleOCR หลัง HTTP สำหรับ slip-reader — ไม่มีโค้ด Python ที่เขียนเอง ใช้ `paddlex --serve` กับ `OCR.yaml` ที่ pin โมเดล

## สัญญา HTTP (จาก schema ของ PaddleX 3.7.2 — ยืนยันกับ container จริงแล้ว: ⬜ ยัง)

```
POST /ocr
{"file": "<base64 ของรูป>", "fileType": 1}          # 1 = รูปภาพ, 0 = PDF

200 {"logId": "...", "errorCode": 0, "errorMsg": "Success",
     "result": {"ocrResults": [{"prunedResult": {"rec_texts": ["บรรทัด1", "บรรทัด2", ...], ...}}], "dataInfo": {...}}}

GET /health → 200
```

client ฝั่ง Go อยู่ที่ `ocr/paddle.go` — ถ้า upstream เปลี่ยน schema ให้แก้ตรงนั้นที่เดียว

## ทดสอบด้วยมือ

```sh
docker compose up -d --build
curl -s -X POST localhost:8866/ocr -H 'content-type: application/json' \
  -d "{\"file\":\"$(base64 -i slip.png)\",\"fileType\":1}" | head -c 600
```

## หน่วยความจำ

mobile pipeline พีค ~2.4 GB ที่ 4 thread (วัดตอน spike) — `docker-compose.yml` ตั้ง `mem_limit: 4g`
ห้ามสลับไป `PP-OCRv5_server_det`: ช้ากว่า ~50 เท่าบน CPU และเคยกินหน่วยความจำจนระบบฆ่า process
