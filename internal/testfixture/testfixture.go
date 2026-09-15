// Package testfixture โหลด fixture จาก testdata/ocr ให้เทสต์ของหลาย package ใช้ร่วมกัน
package testfixture

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// Case คือ fixture หนึ่งชุด: ข้อความ OCR ดิบ กับ JSON ของ Document ที่คาดไว้
type Case struct {
	Name     string   // เช่น "kbank/merchant"
	Issuer   string   // ชื่อโฟลเดอร์ = issuer ที่ต้องเดาให้ถูก
	RawLines []string // ยังไม่ผ่าน textnorm.CleanLines — ให้แต่ละเทสต์ตัดสินเอง
	WantJSON []byte   // ผู้เรียก unmarshal เป็น type ของตัวเอง เพื่อไม่ให้ package นี้ต้อง import slipreader
}

// Dir คืน path ของ testdata/ocr โดยอิงจากตำแหน่งไฟล์นี้ ไม่ใช่ working directory
// เพราะ go test เปลี่ยน cwd ไปที่ package ที่กำลังเทสต์ ทำให้ path สัมพัทธ์ต่างกันทุก package
func Dir() string {
	_, thisFile, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(thisFile), "..", "..", "testdata", "ocr")
}

// Load อ่านทุก case ใน testdata/ocr เรียงตามชื่อ
func Load(t *testing.T) []Case {
	t.Helper()
	txtFiles, err := filepath.Glob(filepath.Join(Dir(), "*", "*.txt"))
	if err != nil {
		t.Fatalf("หา fixture ไม่ได้: %v", err)
	}
	if len(txtFiles) == 0 {
		t.Fatalf("ไม่มี fixture ใน %s", Dir())
	}

	cases := make([]Case, 0, len(txtFiles))
	for _, txtPath := range txtFiles {
		raw, err := os.ReadFile(txtPath)
		if err != nil {
			t.Fatalf("อ่าน %s: %v", txtPath, err)
		}
		want, err := os.ReadFile(strings.TrimSuffix(txtPath, ".txt") + ".json")
		if err != nil {
			t.Fatalf("อ่าน JSON คู่ของ %s: %v", txtPath, err)
		}
		issuer := filepath.Base(filepath.Dir(txtPath))
		name := issuer + "/" + strings.TrimSuffix(filepath.Base(txtPath), ".txt")
		cases = append(cases, Case{
			Name:     name,
			Issuer:   issuer,
			RawLines: strings.Split(strings.TrimRight(string(raw), "\n"), "\n"),
			WantJSON: want,
		})
	}
	return cases
}
