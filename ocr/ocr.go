// Package ocr คุยกับ OCR engine ที่อยู่ข้างนอก (container หลัง HTTP)
//
// ส่วนอื่นของไลบรารีเห็นแค่ interface Engine — ไม่รู้ว่าข้างหลังคือ PaddleOCR
// จึงสลับเป็นตัวอื่น หรือใช้ Fake ในเทสต์ได้โดยไม่แตะตัวแกะฟิลด์เลย
package ocr

import "context"

// Engine อ่านข้อความจากรูป คืนทีละบรรทัดเรียงจากบนลงล่าง
type Engine interface {
	Recognize(ctx context.Context, image []byte) ([]string, error)
}
