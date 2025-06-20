package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Cell states
const (
	UNKNOWN = 0
	OFF     = 1
	ON      = -1
)

// Errors
var (
	ErrInvalid    = errors.New("invalid puzzle")
	ErrImpossible = errors.New("impossible to solve")
)

// Puzzle represents a Picross puzzle
type Puzzle struct {
	height      int
	width       int
	grid        [][]int
	transposed  bool
	widthNow    int
	estimates   []int
	changed     []bool
	guess       int
	logging     bool
	hints       [][]int
	rHints      [][]int // row hints
	cHints      [][]int // column hints
}

// Helper functions
func sum(nums []int) int {
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}

func factorial(n int) int {
	result := 1
	for i := 1; i <= n; i++ {
		result *= i
	}
	return result
}

func combination(n, k int) int {
	if k > n || k < 0 {
		return 0
	}
	return factorial(n) / (factorial(k) * factorial(n-k))
}

func max(nums []int) int {
	if len(nums) == 0 {
		return 0
	}
	maxVal := nums[0]
	for _, n := range nums[1:] {
		if n > maxVal {
			maxVal = n
		}
	}
	return maxVal
}

func maxInts(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// cellToString converts a cell value to its string representation
func cellToString(cell int) string {
	switch cell {
	case UNKNOWN:
		return "　"
	case OFF:
		return "×"
	case ON:
		return "■"
	default:
		return "?"
	}
}

// NewPuzzle creates a new puzzle from input lines
func NewPuzzle(lines []string, logging bool) (*Puzzle, error) {
	if len(lines) == 0 {
		return nil, ErrInvalid
	}

	// Parse dimensions
	parts := strings.FieldsFunc(lines[0], func(r rune) bool {
		return r == ',' || r == ' '
	})
	if len(parts) != 2 {
		return nil, ErrInvalid
	}

	height, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, ErrInvalid
	}
	width, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, ErrInvalid
	}

	// Initialize grid
	grid := make([][]int, height)
	for i := range grid {
		grid[i] = make([]int, width)
		// Initialize all cells to UNKNOWN
		for j := range grid[i] {
			grid[i][j] = UNKNOWN
		}
	}

	p := &Puzzle{
		height:    height,
		width:     width,
		grid:      grid,
		transposed: false,
		widthNow:  width,
		estimates: make([]int, height+width),
		changed:   make([]bool, height+width),
		guess:     0,
		logging:   logging,
	}

	// Parse hints
	hints := make([][]int, 0, height+width)
	for i := 1; i < len(lines); i++ {
		if lines[i] == "" {
			continue
		}
		parts := strings.FieldsFunc(lines[i], func(r rune) bool {
			return r == ',' || r == ' '
		})
		hint := make([]int, len(parts))
		for j, part := range parts {
			hint[j], err = strconv.Atoi(part)
			if err != nil {
				return nil, ErrInvalid
			}
		}
		hints = append(hints, hint)
	}

	if len(hints) != height+width {
		return nil, fmt.Errorf("盤面の大きさとヒントの数が一致しません。got %d hints, expected %d", len(hints), height+width)
	}

	p.rHints = hints[0:height]
	p.cHints = hints[height:height+width]

	// Validate hints
	rSum := 0
	for _, hint := range p.rHints {
		rSum += sum(hint)
	}
	cSum := 0
	for _, hint := range p.cHints {
		cSum += sum(hint)
	}
	if rSum != cSum {
		return nil, fmt.Errorf("横のヒントの合計と縦のヒントの合計が一致しません。rows: %d, cols: %d", rSum, cSum)
	}

	// Process hints (simplified for now)
	p.hints = hints

	// Initialize changed flags
	for i := range p.changed {
		p.changed[i] = true
	}

	return p, nil
}

// String returns a string representation of the puzzle
func (p *Puzzle) String() string {
	var buf strings.Builder

	// Find max hint sizes
	rhMax := 0
	for _, hint := range p.rHints {
		if len(hint) > rhMax {
			rhMax = len(hint)
		}
	}
	chMax := 0
	for _, hint := range p.cHints {
		if len(hint) > chMax {
			chMax = len(hint)
		}
	}

	// Print column hints
	for n := chMax - 1; n >= 0; n-- {
		// Print padding
		for i := 0; i < rhMax; i++ {
			buf.WriteString("　")
		}
		buf.WriteString("｜")

		// Print column hint values
		for _, hint := range p.cHints {
			if n < len(hint) {
				fmt.Fprintf(&buf, "%2d", hint[len(hint)-1-n]%100)
			} else {
				buf.WriteString("　")
			}
		}
		buf.WriteString("｜\n")
	}

	// Print separator
	for i := 0; i < rhMax; i++ {
		buf.WriteString("--")
	}
	buf.WriteString("＋")
	for i := 0; i < p.width; i++ {
		buf.WriteString("--")
	}
	buf.WriteString("＋\n")

	// Print rows with hints
	for row, hint := range p.rHints {
		// Print row hints with padding
		padding := rhMax - len(hint)
		for i := 0; i < padding; i++ {
			buf.WriteString("　")
		}
		for _, h := range hint {
			fmt.Fprintf(&buf, "%2d", h%100)
		}
		buf.WriteString("｜")

		// Print cells
		for col := 0; col < p.width; col++ {
			buf.WriteString(cellToString(p.grid[row][col]))
		}
		buf.WriteString("｜\n")
	}

	// Print bottom separator
	for i := 0; i < rhMax; i++ {
		buf.WriteString("--")
	}
	buf.WriteString("＋")
	for i := 0; i < p.width; i++ {
		buf.WriteString("--")
	}
	buf.WriteString("＋")

	return buf.String()
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: picross <filename>")
		os.Exit(1)
	}

	// Read file
	file, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	// Read lines
	scanner := bufio.NewScanner(file)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	// Create puzzle
	logging := len(os.Args) > 2 && os.Args[2] == "-v"
	puzzle, err := NewPuzzle(lines, logging)
	if err != nil {
		fmt.Printf("Error creating puzzle: %v\n", err)
		os.Exit(1)
	}

	// Display initial puzzle
	fmt.Println(puzzle)
}