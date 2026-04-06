package vizualization

import (
	"fmt"
	"math"
	"math/cmplx"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ── FFT (Cooley-Tukey, radix-2) ───────────────────────────────────────────────

func fft(x []complex128) []complex128 {
	n := len(x)
	if n <= 1 {
		return x
	}
	even := make([]complex128, n/2)
	odd := make([]complex128, n/2)
	for i := 0; i < n/2; i++ {
		even[i] = x[2*i]
		odd[i] = x[2*i+1]
	}
	even = fft(even)
	odd = fft(odd)

	result := make([]complex128, n)
	for k := 0; k < n/2; k++ {
		t := cmplx.Rect(1, -2*math.Pi*float64(k)/float64(n)) * odd[k]
		result[k] = even[k] + t
		result[k+n/2] = even[k] - t
	}
	return result
}

func nextPow2(n int) int {
	p := 1
	for p < n {
		p <<= 1
	}
	return p
}

// ── Frequency band analysis ──────────────────────────────────────────────────

const numBands = 32

// Attempt to create logarithmically-spaced bands from ~60 Hz to ~16 kHz
func bandEdges(sampleRate uint32) []float64 {
	lowFreq := 60.0
	hiFreq := math.Min(16000, float64(sampleRate)/2)
	edges := make([]float64, numBands+1)
	logLow := math.Log2(lowFreq)
	logHi := math.Log2(hiFreq)
	for i := 0; i <= numBands; i++ {
		edges[i] = math.Pow(2, logLow+(logHi-logLow)*float64(i)/float64(numBands))
	}
	return edges
}

func analyze(samples []float64, sampleRate uint32) []float64 {
	n := nextPow2(len(samples))
	buf := make([]complex128, n)
	// Apply Hann window
	for i, s := range samples {
		w := 0.5 * (1 - math.Cos(2*math.Pi*float64(i)/float64(len(samples)-1)))
		buf[i] = complex(s*w, 0)
	}

	spectrum := fft(buf)
	freqRes := float64(sampleRate) / float64(n)
	edges := bandEdges(sampleRate)

	bands := make([]float64, numBands)
	for b := 0; b < numBands; b++ {
		loIdx := int(edges[b] / freqRes)
		hiIdx := int(edges[b+1] / freqRes)
		if loIdx < 0 {
			loIdx = 0
		}
		if hiIdx >= n/2 {
			hiIdx = n/2 - 1
		}
		if hiIdx < loIdx {
			hiIdx = loIdx
		}
		var sum float64
		count := 0
		for i := loIdx; i <= hiIdx; i++ {
			mag := cmplx.Abs(spectrum[i]) / float64(n)
			sum += mag
			count++
		}
		if count > 0 {
			bands[b] = sum / float64(count)
		}
	}
	return bands
}

// ── Bubbletea model ──────────────────────────────────────────────────────────

type tickMsg time.Time

type model struct {
	bands       []float64 // current frequency magnitudes
	peak        []float64 // peak hold per band
	peakDecay   []float64 // velocity for peak fall
	elapsed     float64   // seconds into the track
	windowSize  int       // FFT window in samples
	hopSize     int       // advance per tick in samples
	width       int
	height      int
	fps         int
	quitting    bool
	filename    string
	smoothBands []float64 // smoothed display values
}

func initialModel() model {
	fps := 30
	windowSamples := int(float64(song.sampleRate) * 0.05) // 50ms window
	hopSamples := int(audio.sampleRate) / fps

	m := model{
		bands:       make([]float64, numBands),
		peak:        make([]float64, numBands),
		peakDecay:   make([]float64, numBands),
		smoothBands: make([]float64, numBands),
		windowSize:  windowSamples,
		hopSize:     hopSamples,
		width:       120,
		height:      28,
		fps:         fps,
		filename:    filename,
	}
	return m
}

func tickCmd(fps int) tea.Cmd {
	return tea.Tick(time.Second/time.Duration(fps), func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) Init() tea.Cmd {
	return tickCmd(m.fps)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height - 6 // reserve for header/footer
		if m.height < 4 {
			m.height = 4
		}
	case tickMsg:
		if m.elapsed >= m.audio.totalSecs {
			m.quitting = true
			return m, tea.Quit
		}
		// Extract window of samples at current position
		startSample := int(m.elapsed * float64(m.audio.sampleRate))
		endSample := startSample + m.windowSize
		if endSample > len(m.audio.samples) {
			endSample = len(m.audio.samples)
		}
		if startSample < endSample {
			window := m.audio.samples[startSample:endSample]
			m.bands = analyze(window, m.audio.sampleRate)
		}
		// Smooth & peak hold
		for i := 0; i < numBands; i++ {
			// Exponential smoothing
			alpha := 0.35
			m.smoothBands[i] = alpha*m.bands[i] + (1-alpha)*m.smoothBands[i]

			// Peak hold with gravity
			if m.smoothBands[i] > m.peak[i] {
				m.peak[i] = m.smoothBands[i]
				m.peakDecay[i] = 0
			} else {
				m.peakDecay[i] += 0.0004
				m.peak[i] -= m.peakDecay[i]
				if m.peak[i] < 0 {
					m.peak[i] = 0
				}
			}
		}

		m.elapsed += float64(m.hopSize) / float64(m.audio.sampleRate)
		return m, tickCmd(m.fps)
	}
	return m, nil
}

// ── Rendering ─────────────────────────────────────────────────────────────────

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF79C6")).
			PaddingLeft(1)

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6272A4")).
			PaddingLeft(1)

	timeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#50FA7B")).
			Bold(true)

	brandStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#BD93F9")).
			Bold(true)
)

// Gradient from deep blue → cyan → green → yellow → red → magenta
var barColors = []string{
	"#6272A4", "#5B7FBA", "#4ECDC4", "#50FA7B",
	"#69FF94", "#F1FA8C", "#FFB86C", "#FF6E6E",
	"#FF5555", "#FF79C6",
}

func colorForHeight(row, maxRows int) lipgloss.Style {
	idx := int(float64(maxRows-1-row) / float64(maxRows) * float64(len(barColors)-1))
	if idx < 0 {
		idx = 0
	}
	if idx >= len(barColors) {
		idx = len(barColors) - 1
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color(barColors[idx]))
}

func peakColor() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#F8F8F2")).Bold(true)
}

func (m model) View() string {
	if m.quitting {
		return "\n  Done! 🎵\n\n"
	}

	var b strings.Builder

	// Header
	b.WriteString("\n")
	b.WriteString(brandStyle.Render("  ♫ FLAC-VIZ"))
	b.WriteString(titleStyle.Render(fmt.Sprintf(" │ %s", m.filename)))
	b.WriteString("\n")

	// Time bar
	elapsed := time.Duration(m.elapsed * float64(time.Second))
	total := time.Duration(m.audio.totalSecs * float64(time.Second))
	progress := m.elapsed / m.audio.totalSecs
	barW := m.width - 22
	if barW < 10 {
		barW = 10
	}
	filled := int(progress * float64(barW))
	if filled > barW {
		filled = barW
	}

	timeStr := fmt.Sprintf("  %s / %s ",
		formatDuration(elapsed),
		formatDuration(total))
	b.WriteString(timeStyle.Render(timeStr))

	progressBar := lipgloss.NewStyle().Foreground(lipgloss.Color("#BD93F9")).Render(strings.Repeat("━", filled))
	progressBar += lipgloss.NewStyle().Foreground(lipgloss.Color("#44475A")).Render(strings.Repeat("─", barW-filled))
	b.WriteString(progressBar)
	b.WriteString("\n\n")

	// Spectrum bars
	maxH := m.height
	if maxH < 4 {
		maxH = 4
	}

	// Find max magnitude for normalization (with floor to avoid division by 0)
	maxMag := 0.001
	for _, v := range m.smoothBands {
		if v > maxMag {
			maxMag = v
		}
	}
	// Use a log scale for better visual dynamic range
	barWidth := (m.width - 4) / numBands
	if barWidth < 1 {
		barWidth = 1
	}
	gap := 0
	if barWidth > 2 {
		gap = 1
		barWidth--
	}
	barChar := "█"
	if barWidth == 1 {
		barChar = "▌"
	}

	for row := 0; row < maxH; row++ {
		b.WriteString("  ")
		threshold := 1.0 - float64(row)/float64(maxH)
		for band := 0; band < numBands; band++ {
			// Normalize to 0..1 with log scaling
			normalized := m.smoothBands[band] / maxMag
			normalized = math.Log1p(normalized*9) / math.Log1p(9) // log compression
			peakNorm := m.peak[band] / maxMag
			peakNorm = math.Log1p(peakNorm*9) / math.Log1p(9)

			peakRow := 1.0 - peakNorm
			if row == int(peakRow*float64(maxH)) && peakNorm > 0.01 {
				b.WriteString(peakColor().Render(strings.Repeat("▔", barWidth)))
			} else if normalized >= threshold {
				style := colorForHeight(row, maxH)
				b.WriteString(style.Render(strings.Repeat(barChar, barWidth)))
			} else {
				b.WriteString(strings.Repeat(" ", barWidth))
			}
			if gap > 0 {
				b.WriteString(" ")
			}
		}
		b.WriteString("\n")
	}

	// Frequency labels
	b.WriteString("  ")
	edges := bandEdges(m.audio.sampleRate)
	labelPositions := []int{0, numBands / 4, numBands / 2, 3 * numBands / 4, numBands - 1}
	labelLine := make([]byte, numBands*(barWidth+gap)+1)
	for i := range labelLine {
		labelLine[i] = ' '
	}
	for _, pos := range labelPositions {
		freq := (edges[pos] + edges[pos+1]) / 2
		var label string
		if freq >= 1000 {
			label = fmt.Sprintf("%.1fk", freq/1000)
		} else {
			label = fmt.Sprintf("%.0f", freq)
		}
		col := pos * (barWidth + gap)
		if col+len(label) <= len(labelLine) {
			copy(labelLine[col:], label)
		}
	}
	b.WriteString(footerStyle.Render(string(labelLine)))
	b.WriteString("\n")

	// Footer
	b.WriteString(footerStyle.Render(fmt.Sprintf("  %d Hz │ %d bands │ press q to quit",
		m.audio.sampleRate, numBands)))
	b.WriteString("\n")

	return b.String()
}

func formatDuration(d time.Duration) string {
	m := int(d.Minutes())
	s := int(d.Seconds()) % 60
	return fmt.Sprintf("%d:%02d", m, s)
}
