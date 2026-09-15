// accuracy รันสลิปจริงทั้งคลังผ่าน slip-reader แล้วพิมพ์ความแม่นรายฟิลด์ราย issuer
//
//	accuracy -slips testdata/slips-raw -labels testdata/slips-raw/labels.csv [-ocr URL] [-cache DIR] [-min 0.95]
//
// ผล OCR ดิบของแต่ละรูปถูก cache ไว้ (ค่าตั้งต้น testdata/ocr-cache) เพื่อให้รันซ้ำตอนแก้ตัวแกะฟิลด์
// ใช้เวลาเป็นวินาที ไม่ใช่นาที — ลบ cache เมื่อเปลี่ยนโมเดล OCR
// exit code 1 ถ้าฟิลด์ใดแม่นต่ำกว่า -min เพื่อใช้ใน CI ได้
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"

	slipreader "github.com/Muegoo/slip-reader"
	"github.com/Muegoo/slip-reader/internal/textnorm"
	"github.com/Muegoo/slip-reader/ocr"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "accuracy:", err)
		os.Exit(1)
	}
}

var errBelowTarget = errors.New("ความแม่นต่ำกว่าเป้า")

func run(args []string, out io.Writer) error {
	flags := flag.NewFlagSet("accuracy", flag.ContinueOnError)
	slipsDir := flags.String("slips", "testdata/slips-raw", "โฟลเดอร์รูปสลิปจริง")
	labelsPath := flags.String("labels", "testdata/slips-raw/labels.csv", "ไฟล์ ground truth")
	ocrURL := flags.String("ocr", envOr("SLIPREADER_OCR_URL", "http://localhost:8866"), "base URL ของ ocr-service")
	cacheDir := flags.String("cache", "testdata/ocr-cache", "โฟลเดอร์เก็บผล OCR ดิบ")
	minRate := flags.Float64("min", 0.95, "ความแม่นต่ำสุดที่ยอมรับต่อฟิลด์ (0-1)")
	if err := flags.Parse(args); err != nil {
		return err
	}

	labelsFile, err := os.Open(*labelsPath)
	if err != nil {
		return fmt.Errorf("เปิด labels: %w", err)
	}
	defer labelsFile.Close()
	labels, err := readLabels(labelsFile)
	if err != nil {
		return err
	}

	engine := ocr.NewPaddle(*ocrURL)
	board := newScoreboard()
	for _, label := range labels {
		raw, err := recognizeWithCache(engine, *slipsDir, *cacheDir, label.File)
		if err != nil {
			return fmt.Errorf("%s: %w", label.File, err)
		}
		// อ่านผ่าน pipeline จริงทุกขั้น ยกเว้น OCR ที่ป้อนผลจาก cache/engine เข้าไปแทน
		doc, err := slipreader.New(&ocr.Fake{Lines: raw}).Read(context.Background(), nil)
		if err != nil {
			return fmt.Errorf("%s: %w", label.File, err)
		}
		board.record(label, doc)
	}

	board.print(out)
	if board.belowTarget(*minRate) {
		return fmt.Errorf("%w %.0f%%", errBelowTarget, *minRate*100)
	}
	return nil
}

// recognizeWithCache คืนบรรทัด OCR ของรูป จาก cache ถ้ามี ไม่งั้นเรียก engine แล้วเก็บลง cache
func recognizeWithCache(engine ocr.Engine, slipsDir, cacheDir, file string) ([]string, error) {
	cachePath := filepath.Join(cacheDir, file+".json")
	if data, err := os.ReadFile(cachePath); err == nil {
		var lines []string
		if err := json.Unmarshal(data, &lines); err == nil {
			return lines, nil
		}
	}

	image, err := os.ReadFile(filepath.Join(slipsDir, file))
	if err != nil {
		return nil, fmt.Errorf("อ่านรูป: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	lines, err := engine.Recognize(ctx, image)
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return nil, fmt.Errorf("สร้าง cache: %w", err)
	}
	data, _ := json.Marshal(lines)
	if err := os.WriteFile(cachePath, data, 0o644); err != nil {
		return nil, fmt.Errorf("เขียน cache: %w", err)
	}
	return lines, nil
}

// ---- การนับคะแนน ----

type tally struct {
	total, issuer, amount, counterparty, date int
}

type miss struct {
	file, field, got, want string
}

type scoreboard struct {
	byIssuer  map[string]*tally
	misses    []miss
	negatives int // รูปที่ไม่ใช่สลิป
	hallucin  int // รูปที่ไม่ใช่สลิปแต่ระบบอ่านยอดออกมา
}

func newScoreboard() *scoreboard {
	return &scoreboard{byIssuer: map[string]*tally{}}
}

func (s *scoreboard) record(label Label, doc *slipreader.Document) {
	if !label.IsSlip {
		s.negatives++
		if doc.Confidence.Amount != slipreader.LevelMissing {
			s.hallucin++
			s.misses = append(s.misses, miss{label.File, "ไม่ใช่สลิปแต่อ่านยอด", fmt.Sprint(doc.AmountSatang), "missing"})
		}
		return
	}

	t := s.byIssuer[label.Issuer]
	if t == nil {
		t = &tally{}
		s.byIssuer[label.Issuer] = t
	}
	t.total++

	if string(doc.Issuer) == label.Issuer {
		t.issuer++
	} else {
		s.misses = append(s.misses, miss{label.File, "issuer", string(doc.Issuer), label.Issuer})
	}
	if doc.AmountSatang == int64(label.AmountSatang) && doc.Confidence.Amount != slipreader.LevelMissing {
		t.amount++
	} else {
		s.misses = append(s.misses, miss{label.File, "ยอด", fmt.Sprint(doc.AmountSatang), fmt.Sprint(label.AmountSatang)})
	}
	if textnorm.Similar(doc.Counterparty, label.Counterparty) && doc.Counterparty != "" {
		t.counterparty++
	} else {
		s.misses = append(s.misses, miss{label.File, "คู่ค้า", doc.Counterparty, label.Counterparty})
	}
	if sameDay(doc.OccurredAt, label.Date) {
		t.date++
	} else {
		s.misses = append(s.misses, miss{label.File, "วันที่", doc.OccurredAt.Format("2006-01-02"), label.Date.Format("2006-01-02")})
	}
}

func sameDay(got time.Time, want time.Time) bool {
	if got.IsZero() {
		return false
	}
	return got.Format("2006-01-02") == want.Format("2006-01-02")
}

func (s *scoreboard) sum() tally {
	var all tally
	for _, t := range s.byIssuer {
		all.total += t.total
		all.issuer += t.issuer
		all.amount += t.amount
		all.counterparty += t.counterparty
		all.date += t.date
	}
	return all
}

func (s *scoreboard) belowTarget(min float64) bool {
	all := s.sum()
	if all.total == 0 {
		return true
	}
	rate := func(n int) float64 { return float64(n) / float64(all.total) }
	return rate(all.issuer) < min || rate(all.amount) < min || rate(all.counterparty) < min || rate(all.date) < min
}

func (s *scoreboard) print(out io.Writer) {
	issuers := make([]string, 0, len(s.byIssuer))
	for name := range s.byIssuer {
		issuers = append(issuers, name)
	}
	sort.Strings(issuers)

	fmt.Fprintln(out, "| issuer | ใบ | issuer ถูก | ยอด | คู่ค้า | วันที่ |")
	fmt.Fprintln(out, "|---|---|---|---|---|---|")
	for _, name := range issuers {
		t := s.byIssuer[name]
		fmt.Fprintf(out, "| %s | %d | %d/%d | %d/%d | %d/%d | %d/%d |\n",
			name, t.total, t.issuer, t.total, t.amount, t.total, t.counterparty, t.total, t.date, t.total)
	}
	all := s.sum()
	fmt.Fprintf(out, "| **รวม** | %d | **%d/%d** | **%d/%d** | **%d/%d** | **%d/%d** |\n",
		all.total, all.issuer, all.total, all.amount, all.total, all.counterparty, all.total, all.date, all.total)
	fmt.Fprintf(out, "\nรูปที่ไม่ใช่สลิป %d ใบ — ระบบอ่านยอดออกมาผิด ๆ %d ใบ\n", s.negatives, s.hallucin)

	if len(s.misses) > 0 {
		fmt.Fprintf(out, "\nพลาด (%d):\n", len(s.misses))
		for _, m := range s.misses {
			fmt.Fprintf(out, "- %s — %s: ได้ %q ต้องการ %q\n", m.file, m.field, m.got, m.want)
		}
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
