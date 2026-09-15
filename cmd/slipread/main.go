// slipread อ่านสลิปหนึ่งรูปแล้วพิมพ์ Document เป็น JSON — ไว้ลองด้วยมือและดีบักตัวแกะฟิลด์
//
//	slipread [-ocr http://localhost:8866] [-raw] image.jpg
//
// ค่าตั้งต้นของ -ocr อ่านจาก env SLIPREADER_OCR_URL ถ้าตั้งไว้
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	slipreader "github.com/Muegoo/slip-reader"
	"github.com/Muegoo/slip-reader/ocr"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "slipread:", err)
		os.Exit(1)
	}
}

func run(args []string, out *os.File) error {
	flags := flag.NewFlagSet("slipread", flag.ContinueOnError)
	ocrURL := flags.String("ocr", envOr("SLIPREADER_OCR_URL", "http://localhost:8866"), "base URL ของ ocr-service")
	showRaw := flags.Bool("raw", false, "พิมพ์ raw_lines จาก OCR ด้วย (ค่าตั้งต้นตัดออกให้ output สั้น)")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 1 {
		return fmt.Errorf("ต้องระบุไฟล์รูป 1 ไฟล์ — usage: slipread [-ocr URL] [-raw] image.jpg")
	}

	image, err := os.ReadFile(flags.Arg(0))
	if err != nil {
		return fmt.Errorf("อ่านรูป: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	doc, err := slipreader.New(ocr.NewPaddle(*ocrURL)).Read(ctx, image)
	if err != nil {
		return err
	}
	if !*showRaw {
		doc.RawLines = nil
	}

	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)
	return encoder.Encode(doc)
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
