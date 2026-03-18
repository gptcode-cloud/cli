package autonomous

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

type Spinner struct {
	message string
	done    chan bool
	mu      sync.Mutex
	running bool
	frames  []string
}

var defaultFrames = []string{
	"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏",
}

func NewSpinner(message string) *Spinner {
	return &Spinner{
		message: message,
		done:    make(chan bool),
		frames:  defaultFrames,
	}
}

func (s *Spinner) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.mu.Unlock()

	go func() {
		i := 0
		for {
			select {
			case <-s.done:
				return
			default:
				s.mu.Lock()
				frame := s.frames[i%len(s.frames)]
				i++
				s.mu.Unlock()

				fmt.Printf("\r%s %s", frame, s.message)
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()
}

func (s *Spinner) Stop() {
	s.mu.Lock()
	s.running = false
	s.mu.Unlock()

	s.done <- true
	fmt.Printf("\r%s\n", strings.Repeat(" ", len(s.message)+2))
}

func (s *Spinner) SetMessage(msg string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.message = msg
}

type ProgressBar struct {
	width   int
	total   int
	current int
	prefix  string
	mu      sync.Mutex
}

func NewProgressBar(total int) *ProgressBar {
	return &ProgressBar{
		width:   40,
		total:   total,
		current: 0,
	}
}

func (pb *ProgressBar) SetPrefix(prefix string) {
	pb.mu.Lock()
	defer pb.mu.Unlock()
	pb.prefix = prefix
}

func (pb *ProgressBar) Increment() {
	pb.mu.Lock()
	defer pb.mu.Unlock()
	pb.current++
	pb.draw()
}

func (pb *ProgressBar) Set(n int) {
	pb.mu.Lock()
	defer pb.mu.Unlock()
	pb.current = n
	pb.draw()
}

func (pb *ProgressBar) draw() {
	if pb.total == 0 {
		return
	}

	filled := (pb.width * pb.current) / pb.total
	empty := pb.width - filled

	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)
	percent := (pb.current * 100) / pb.total

	prefix := pb.prefix
	if prefix == "" {
		prefix = "Progress"
	}

	fmt.Printf("\r%s [%s] %d%%", prefix, bar, percent)
	if pb.current >= pb.total {
		fmt.Println()
	}
}

type Color int

const (
	Reset Color = iota
	Bold
	Red
	Green
	Yellow
	Blue
	Magenta
	Cyan
	White
)

var colorCodes = map[Color]string{
	Reset:   "\033[0m",
	Bold:    "\033[1m",
	Red:     "\033[31m",
	Green:   "\033[32m",
	Yellow:  "\033[33m",
	Blue:    "\033[34m",
	Magenta: "\033[35m",
	Cyan:    "\033[36m",
	White:   "\033[37m",
}

func Colorize(text string, colors ...Color) string {
	var codes strings.Builder
	for _, c := range colors {
		codes.WriteString(colorCodes[c])
	}
	return codes.String() + text + colorCodes[Reset]
}

func Success(text string) {
	fmt.Println(Colorize("✓ "+text, Green, Bold))
}

func Error(text string) {
	fmt.Println(Colorize("✗ "+text, Red, Bold))
}

func Warning(text string) {
	fmt.Println(Colorize("⚠ "+text, Yellow, Bold))
}

func Info(text string) {
	fmt.Println(Colorize("ℹ "+text, Cyan))
}

func Debug(text string) {
	if os.Getenv("GPTCODE_DEBUG") == "1" {
		fmt.Println(Colorize("[DEBUG] "+text, White))
	}
}

type Printer struct {
	quiet   bool
	verbose bool
	colors  bool
}

func NewPrinter() *Printer {
	return &Printer{
		colors: true,
	}
}

func (p *Printer) SetQuiet(quiet bool) {
	p.quiet = quiet
}

func (p *Printer) SetVerbose(verbose bool) {
	p.verbose = verbose
}

func (p *Printer) SetColors(colors bool) {
	p.colors = colors
}

func (p *Printer) Print(text string, colors ...Color) {
	if p.quiet {
		return
	}
	if p.colors && len(colors) > 0 {
		fmt.Print(Colorize(text, colors...))
	} else {
		fmt.Print(text)
	}
}

func (p *Printer) Println(text string, colors ...Color) {
	if p.quiet {
		return
	}
	if p.colors && len(colors) > 0 {
		fmt.Println(Colorize(text, colors...))
	} else {
		fmt.Println(text)
	}
}

func (p *Printer) Debug(text string) {
	if p.verbose || os.Getenv("GPTCODE_DEBUG") == "1" {
		p.Println("[DEBUG] "+text, White)
	}
}

func (p *Printer) Success(text string) {
	p.Println("✓ "+text, Green, Bold)
}

func (p *Printer) Error(text string) {
	p.Println("✗ "+text, Red, Bold)
}

func (p *Printer) Warning(text string) {
	p.Println("⚠ "+text, Yellow, Bold)
}

func (p *Printer) Info(text string) {
	p.Println("ℹ "+text, Cyan)
}

func (p *Printer) Header(text string) {
	if p.quiet {
		return
	}
	fmt.Println()
	fmt.Println(Colorize("═══ "+text+" ═══", Bold, Cyan))
}

func (p *Printer) SubHeader(text string) {
	if p.quiet {
		return
	}
	fmt.Println(Colorize("─── "+text+" ───", Bold, White))
}

func (p *Printer) Item(label, value string) {
	if p.quiet {
		return
	}
	fmt.Printf("  %-20s %s\n", label+":", value)
}

func (p *Printer) Separator() {
	if p.quiet {
		return
	}
	fmt.Println(strings.Repeat("─", 40))
}

type Table struct {
	headers []string
	rows    [][]string
	widths  []int
}

func NewTable(headers ...string) *Table {
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	return &Table{
		headers: headers,
		widths:  widths,
	}
}

func (t *Table) AddRow(values ...string) {
	t.rows = append(t.rows, values)
	for i, v := range values {
		if len(v) > t.widths[i] {
			t.widths[i] = len(v)
		}
	}
}

func (t *Table) Print() {
	// Print header
	for i, h := range t.headers {
		fmt.Printf(" %-*s ", t.widths[i], h)
		if i < len(t.headers)-1 {
			fmt.Print(" │ ")
		}
	}
	fmt.Println()

	// Print separator
	for i, w := range t.widths {
		fmt.Print("─" + strings.Repeat("─", w))
		if i < len(t.widths)-1 {
			fmt.Print("─┼─")
		}
	}
	fmt.Println()

	// Print rows
	for _, row := range t.rows {
		for i, v := range row {
			fmt.Printf(" %-*s ", t.widths[i], v)
			if i < len(row)-1 {
				fmt.Print(" │ ")
			}
		}
		fmt.Println()
	}
}
