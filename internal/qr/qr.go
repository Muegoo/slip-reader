// Package qr ถอด QR code บนสลิป เพื่อเอาเลขอ้างอิงที่แม่นกว่าที่ OCR อ่านได้
//
// สลิปโอนเงินของธนาคารไทยส่วนใหญ่มี mini-QR สำหรับ "สแกนตรวจสอบสลิป" ข้างในเป็นข้อความอ้างอิงรายการ
// ถอดได้ = ได้กุญแจกันบันทึกซ้ำที่ไม่มีตัวเลขเพี้ยน ถอดไม่ได้ = ไม่เป็นไร ใช้ค่าจาก OCR แทน
package qr

import (
	"bytes"
	"image"
	_ "image/jpeg" // ลงทะเบียน decoder — สลิปเป็น JPEG หรือ PNG เท่านั้น
	_ "image/png"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
)

// Decode คืนข้อความใน QR ตัวแรกที่ถอดได้ หรือ false ถ้าไม่มี / รูปเปิดไม่ได้
// ไม่คืน error เพราะการไม่มี QR เป็นเรื่องปกติ (ใบเสร็จ 7-Eleven และ e-wallet บางเจ้าไม่มี)
func Decode(imageBytes []byte) (string, bool) {
	img, _, err := image.Decode(bytes.NewReader(imageBytes))
	if err != nil {
		return "", false
	}
	bitmap, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		return "", false
	}
	// TryHarder ให้ตัวถอดลองหลายมุมและหลายขนาด — QR บนสลิปเล็กและอยู่มุมล่าง
	hints := map[gozxing.DecodeHintType]interface{}{gozxing.DecodeHintType_TRY_HARDER: true}
	result, err := qrcode.NewQRCodeReader().Decode(bitmap, hints)
	if err != nil {
		return "", false
	}
	return result.GetText(), true
}
