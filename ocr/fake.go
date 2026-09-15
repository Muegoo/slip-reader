package ocr

import "context"

// Fake คืนบรรทัดที่กำหนดไว้ล่วงหน้า ใช้ในเทสต์ของ slipreader และตอนลอง CLI โดยไม่ต้องมี container
type Fake struct {
	Lines []string
	Err   error
}

func (f *Fake) Recognize(context.Context, []byte) ([]string, error) {
	if f.Err != nil {
		return nil, f.Err
	}
	return f.Lines, nil
}
