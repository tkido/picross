package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

const (
	UNKNOWN = 0
	OFF     = 1
	ON      = -1
)

var TABLE = map[int]string{
	UNKNOWN: "　",
	OFF:     "×",
	ON:      "■",
}

// Helper functions for math
func sum(arr []int) int {
	s := 0
	for _, v := range arr {
		s += v
	}
	return s
}

func factorial(n int) int {
	if n <= 1 {
		return 1
	}
	r := 1
	for i := 1; i <= n; i++ {
		r *= i
	}
	return r
}

func combination(n, k int) int {
	if k < 0 || k > n {
		return 0
	}
	return factorial(n) / (factorial(k) * factorial(n-k))
}

// Puzzle はピクロスの盤面を表す構造体
type Puzzle struct {
	Height              int
	Width               int
	WidthNow            int
	Grid                [][]int
	Transposed          bool
	Hints               [][]int
	RHints              [][]int
	CHints              [][]int
	OrigRHints          [][]int // 表示用の元のヒント（行）
	OrigCHints          [][]int // 表示用の元のヒント（列）
	Estimates           []int
	Changed             []bool
	GuessPos            int
	Logging             bool
	Writer              io.Writer
	InitRowEstimatesSum int // 初期見積もり（行の合計）
	InitColEstimatesSum int // 初期見積もり（列の合計）
}

func NewPuzzle(lines []string, logging bool) (*Puzzle, error) {
	if len(lines) == 0 {
		return nil, errors.New("empty input")
	}

	// Parse dimensions
	dims := splitInts(lines[0])
	if len(dims) < 2 {
		return nil, errors.New("invalid dimensions")
	}
	height, width := dims[0], dims[1]

	p := &Puzzle{
		Height:   height,
		Width:    width,
		WidthNow: width,

		Estimates: make([]int, height+width),
		Changed:   make([]bool, height+width),
		Logging:   logging,
		Writer:    os.Stdout,
	}

	for i := 0; i < height; i++ {
		p.Grid = append(p.Grid, make([]int, width)) // Use append to initialize correctly if Gener was typo
		p.Grid[i] = make([]int, width)
		for j := 0; j < width; j++ {
			p.Grid[i][j] = UNKNOWN
		}
	}
	// Fix: p.Grid initialization above had a typo in previous thought, simplified here:
	p.Grid = make([][]int, height)
	for i := range p.Grid {
		p.Grid[i] = make([]int, width)
	}

	hintLines := lines[1:]
	if len(hintLines) != height+width {
		return nil, errors.New("number of hints does not match dimensions")
	}

	p.Hints = make([][]int, len(hintLines))
	for i, line := range hintLines {
		p.Hints[i] = splitInts(line)
	}

	p.RHints = p.Hints[:height]
	p.CHints = p.Hints[height:]

	if sumHints(p.RHints) != sumHints(p.CHints) {
		return nil, errors.New("sum of row hints and column hints do not match")
	}

	// 表示用に元のヒントを保存
	p.OrigRHints = make([][]int, len(p.RHints))
	for i, h := range p.RHints {
		p.OrigRHints[i] = make([]int, len(h))
		copy(p.OrigRHints[i], h)
	}
	p.OrigCHints = make([][]int, len(p.CHints))
	for i, h := range p.CHints {
		p.OrigCHints[i] = make([]int, len(h))
		copy(p.OrigCHints[i], h)
	}

	// Initialize estimates and changed flags
	newHints := make([][]int, 0, len(p.Hints))

	// We need to iterate exactly like Ruby: each_line equivalent logic here or just loop
	// In Ruby init:
	/*
	   hints_new = []
	   each_line do |row, index, hint|
	     ...
	   end
	   @hints = hints_new
	*/
	// Since 'each_line' logic depends on hints traversal which is 0..H+W-1
	// And in init, transposed is false.

	for index, hint := range p.Hints {
		// Validation
		if p.WidthNow < sum(hint)+len(hint)-1 {
			return nil, errors.New("hints too large for row/col")
		}

		if len(hint) == 1 && hint[0] == 0 {
			p.Changed[index] = true
			p.Estimates[index] = 1
			// ヒント[0]は全セルOFFを意味する。行ヒントはWidth、列ヒントはHeight（転置後の幅）を使う
			lineLen := p.Width
			if index >= height {
				lineLen = p.Height
			}
			newHints = append(newHints, []int{lineLen})
		} else {
			maxH := 0
			for _, h := range hint {
				if h > maxH {
					maxH = h
				}
			}
			if maxH > p.WidthNow-(sum(hint)+len(hint)-1) {
				p.Changed[index] = true
			}
			p.Estimates[index] = combination(p.WidthNow-sum(hint)+1, len(hint))

			// arr = [0]; hint.each{|n| arr.push(n, 1)}; arr.pop; hints_new.push arr.push(0)
			// This adds explicit 0 gap requirements to the hint array for the recursive solver?
			// e.g. hint [2, 1] becomes [0, 2, 1, 1, 1, 0] ? No.
			// Ruby: arr.push(n, 1) -> [0, 2, 1, 1, 1] -> pop -> [0, 2, 1, 1] -> push(0) -> [0, 2, 1, 1, 0]
			// This structure seems to be: [gap_min, block_size, gap_min, block_size, ..., gap_min]

			arr := []int{0}
			for _, n := range hint {
				arr = append(arr, n, 1)
			}
			if len(arr) > 0 {
				arr = arr[:len(arr)-1] // pop
			}
			arr = append(arr, 0)
			newHints = append(newHints, arr)
		}
	}
	p.Hints = newHints
	// Update RHints/CHints views
	p.RHints = p.Hints[:height]
	p.CHints = p.Hints[height:]

	// 初期見積もりの合計を計算
	for i := 0; i < height; i++ {
		p.InitRowEstimatesSum += p.Estimates[i]
	}
	for i := height; i < height+width; i++ {
		p.InitColEstimatesSum += p.Estimates[i]
	}

	return p, nil
}

func splitInts(s string) []int {
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == ',' || r == ' ' || r == '\t'
	})
	res := make([]int, 0, len(parts))
	for _, p := range parts {
		if p == "" {
			continue
		}
		v, _ := strconv.Atoi(p)
		res = append(res, v)
	}
	return res
}

func sumHints(hints [][]int) int {
	t := 0
	for _, h := range hints {
		t += sum(h)
	}
	return t
}

func (p *Puzzle) String(cursorRow int) string {
	// 表示用の元のヒントを使用
	rHintsDisp := p.OrigRHints
	cHintsDisp := p.OrigCHints
	if p.Transposed {
		rHintsDisp, cHintsDisp = cHintsDisp, rHintsDisp
	}

	rhMax := 0
	for _, h := range rHintsDisp {
		if len(h) > rhMax {
			rhMax = len(h)
		}
	}
	chMax := 0
	for _, h := range cHintsDisp {
		if len(h) > chMax {
			chMax = len(h)
		}
	}

	var sb strings.Builder

	// 列ヘッダー
	for n := chMax - 1; n >= 0; n-- {
		sb.WriteString(strings.Repeat("　", rhMax) + "｜")
		for _, hint := range cHintsDisp {
			if n < len(hint) {
				val := hint[len(hint)-1-n]
				sb.WriteString(fmt.Sprintf("%2d", val%100))
			} else {
				sb.WriteString("　")
			}
		}
		sb.WriteString("｜\n")
	}

	sb.WriteString(strings.Repeat("--", rhMax) + "＋" + strings.Repeat("--", p.WidthNow) + "＋\n")

	for row, hint := range rHintsDisp {
		sb.WriteString(strings.Repeat("　", rhMax-len(hint)))
		for _, n := range hint {
			sb.WriteString(fmt.Sprintf("%2d", n%100))
		}
		sb.WriteString("｜")
		for col := 0; col < p.WidthNow; col++ {
			sb.WriteString(TABLE[p.Grid[row][col]])
		}
		sb.WriteString("｜")
		if row == cursorRow {
			sb.WriteString("<<<<")
		}
		sb.WriteString("\n")
	}
	sb.WriteString(strings.Repeat("--", rhMax) + "＋" + strings.Repeat("--", p.WidthNow) + "＋")
	return sb.String()
}

func (p *Puzzle) Dup() *Puzzle {
	newP := &Puzzle{
		Height:              p.Height,
		Width:               p.Width,
		WidthNow:            p.WidthNow,
		Transposed:          p.Transposed,
		GuessPos:            p.GuessPos,
		Logging:             p.Logging,
		Writer:              p.Writer,
		OrigRHints:          p.OrigRHints,
		OrigCHints:          p.OrigCHints,
		InitRowEstimatesSum: p.InitRowEstimatesSum,
		InitColEstimatesSum: p.InitColEstimatesSum,
	}

	// Deep copy Grid
	newP.Grid = make([][]int, len(p.Grid))
	for i := range p.Grid {
		newP.Grid[i] = make([]int, len(p.Grid[i]))
		copy(newP.Grid[i], p.Grid[i])
	}

	// ヒントの深いコピー
	newP.Hints = make([][]int, len(p.Hints))
	for i, h := range p.Hints {
		newP.Hints[i] = make([]int, len(h))
		copy(newP.Hints[i], h)
	}
	newP.RHints = newP.Hints[:p.Height]
	newP.CHints = newP.Hints[p.Height:]

	newP.Estimates = make([]int, len(p.Estimates))
	copy(newP.Estimates, p.Estimates)

	newP.Changed = make([]bool, len(p.Changed))
	copy(newP.Changed, p.Changed)

	return newP
}

func (p *Puzzle) ChangedAny() bool {
	for _, c := range p.Changed {
		if c {
			return true
		}
	}
	return false
}

func (p *Puzzle) IsSolved() bool {
	for _, row := range p.Grid {
		for _, cell := range row {
			if cell == UNKNOWN {
				return false
			}
		}
	}
	return !p.ChangedAny()
}

func (p *Puzzle) Guess(val int) (*Puzzle, error) {
	for {
		row := p.GuessPos / p.Width
		col := p.GuessPos % p.Width
		if row >= p.Height {
			return p, nil // Should not happen if solving is robust?
		}

		if p.Grid[row][col] == UNKNOWN {
			p.Set(row, col, val)
			p.Changed[row] = true
			return p, nil
		}
		p.GuessPos++
	}
}

func (p *Puzzle) Set(row, col, val int) {
	if p.Grid[row][col] != val {
		p.Grid[row][col] = val
		// Update changed flag
		index := 0
		if p.Transposed {
			index = col
		} else {
			index = p.Height + col
		}
		p.Changed[index] = true
		p.Estimates[index] = p.Estimates[index] * 4 / 5
	}
}

func (p *Puzzle) Transpose() {
	rows := len(p.Grid)
	cols := len(p.Grid[0])
	newGrid := make([][]int, cols)
	for i := 0; i < cols; i++ {
		newGrid[i] = make([]int, rows)
		for j := 0; j < rows; j++ {
			newGrid[i][j] = p.Grid[j][i]
		}
	}
	p.Grid = newGrid
	p.Transposed = !p.Transposed
	if p.Transposed {
		p.WidthNow = p.Height
		// Swap hints
		p.RHints, p.CHints = p.CHints, p.RHints
	} else {
		p.WidthNow = p.Width
		p.RHints, p.CHints = p.CHints, p.RHints
	}
}

func (p *Puzzle) Scan() (*Puzzle, error) {
	// スキャン統計
	totalScanned := 0    // スキャンした行/列の数
	totalCandidates := 0 // 生成した候補の総数

	for p.ChangedAny() {
		for index, hint := range p.Hints {

			var row int
			if p.Transposed {
				row = index - p.OriginalHeight()
			} else {
				row = index
			}

			if index == p.OriginalHeight() && !p.Transposed {
			}

			if p.Changed[index] {
				// p.Height * p.Width is constant total cells.
				limit := p.OriginalHeight() * p.OriginalWidth() / 2
				if p.Estimates[index] < limit {
					candidates, err := p.SolveLine(row, index, hint)
					if err != nil {
						return nil, err
					}
					totalScanned++
					totalCandidates += candidates
					p.Changed[index] = false
					if p.Logging {
						fmt.Fprintf(p.Writer, "★\n%s\n", p.String(row))
					}
				} else {
					p.Estimates[index] = p.Estimates[index] * 4 / 5
				}
			}

			// Logic for transpose trigger
			if index == p.OriginalHeight()-1 {
				p.Transpose()
			}
			if index == len(p.Hints)-1 {
				p.Transpose()
			}
		}
	}

	// スキャン統計をログ出力
	if p.Logging {
		fmt.Fprintf(p.Writer, "スキャン統計: スキャン数=%d, 候補総数=%d\n", totalScanned, totalCandidates)
	}

	return p, nil
}

func (p *Puzzle) OriginalHeight() int {
	// Height is set at init and never swapped in struct?
	// In my Struct `Height`/`Width`, I should keep them constant.
	return p.Height
}

func (p *Puzzle) OriginalWidth() int {
	return p.Width
}

func (p *Puzzle) SolveLine(row, index int, hint []int) (int, error) {
	answers := make([][]int, 0)
	// ヒントのコピーを作成（再帰用）
	hintCopy := make([]int, len(hint))
	copy(hintCopy, hint)

	answers = p.solveRec(row, []int{}, hintCopy, OFF, answers)

	if len(answers) == 0 {
		return 0, errors.New("Impossible")
	}

	// 候補数を記録
	numCandidates := len(answers)

	// RLE形式の回答をフルライン配列に変換
	fullAnswers := make([][]int, len(answers))
	for k, ans := range answers {
		arr := make([]int, 0, p.WidthNow)

		for i, num := range ans {
			val := 0
			if i%2 == 0 {
				val = OFF
			} else {
				val = ON
			} // 1, -1, 1, -1...
			for x := 0; x < num; x++ {
				arr = append(arr, val)
			}
		}
		fullAnswers[k] = arr
	}

	// 各列について全候補の合計から確定セルを判定
	aSize := len(answers)

	for col := 0; col < p.WidthNow; col++ {
		sumVal := 0
		for _, sol := range fullAnswers {
			if col < len(sol) {
				sumVal += sol[col]
			}
		}

		if sumVal == -aSize { // 全候補でON (-1 * count)
			p.Set(row, col, ON)
		} else if sumVal == aSize { // 全候補でOFF (1 * count)
			p.Set(row, col, OFF)
		}
	}
	return numCandidates, nil
}

func (p *Puzzle) solveRec(row int, answer []int, hint []int, val int, results [][]int) [][]int {
	// Ruby _solve
	if len(hint) == 0 {
		if sum(answer) == p.WidthNow {
			// Need to copy answer to store it
			ansCopy := make([]int, len(answer))
			copy(ansCopy, answer)
			results = append(results, ansCopy)
		}
		return results
	}

	// hint check
	// Ruby: `elsif @grid[row][answer.sum, hint.first].include?(-val)`
	// This checks if the segment we are about to place contradicts existing grid.
	// `answer.sum` is current position. `hint.first` is length of next block.
	// If we are placing `val` (e.g. ON or OFF), and grid has `-val` (opposite), it's invalid.
	start := sum(answer)
	length := hint[0]

	// Bounds check?
	if start+length > p.WidthNow {
		// Should be caught by logic below, but careful.
		return results
	}

	conflict := false
	for i := 0; i < length; i++ {
		if start+i < p.WidthNow {
			if p.Grid[row][start+i] == -val {
				conflict = true
				break
			}
		}
	}

	if conflict {
		return results
	}

	// Logic:
	// if (answer.sum + hint.sum) < @width_now and val == OFF
	//   hint_dup = hint.dup; hint_dup[0] += 1
	//   _solve(...)
	// end
	// answer.push hint.shift
	// _solve(..., -val)

	currentHintSum := sum(hint) // hint[0] is included in this
	if (sum(answer)+currentHintSum) < p.WidthNow && val == OFF {
		hintDup := make([]int, len(hint))
		copy(hintDup, hint)
		hintDup[0] += 1

		answerDup := make([]int, len(answer))
		copy(answerDup, answer)

		results = p.solveRec(row, answerDup, hintDup, val, results)
	}

	// shift hint
	shiftVal := hint[0]
	newHint := hint[1:]

	answer = append(answer, shiftVal)
	results = p.solveRec(row, answer, newHint, -val, results)

	return results
}

// Solver は盤面を解くためのソルバー
type Solver struct {
	Logging bool
	Writer  io.Writer // ログ出力先 (デフォルト: os.Stdout)
}

// writer は出力先を返す (未設定の場合は os.Stdout)
func (s *Solver) writer() io.Writer {
	if s.Writer != nil {
		return s.Writer
	}
	return os.Stdout
}

func (s *Solver) Solve(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	var lines []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}

	puzzle, err := NewPuzzle(lines, s.Logging)
	if err != nil {
		return err
	}
	puzzle.Writer = s.writer()

	fmt.Fprintln(s.writer(), puzzle.String(-1))

	// 難易度指標: 初期見積もりをログ出力
	if s.Logging {
		fmt.Fprintf(s.writer(), "初期見積もり: 行合計=%d, 列合計=%d, 総合計=%d\n",
			puzzle.InitRowEstimatesSum, puzzle.InitColEstimatesSum,
			puzzle.InitRowEstimatesSum+puzzle.InitColEstimatesSum)
	}

	// 初期スキャン（制約伝播のみ）
	_, err = puzzle.Scan()
	if err != nil {
		if err.Error() == "Impossible" {
			return errors.New("解がありません。")
		}
		return err
	}

	if puzzle.IsSolved() {
		// 仮定法なしで解けた → 一意解が確定
		if s.Logging {
			fmt.Fprintln(s.writer(), "仮定法の深さ: 0（制約伝播のみで解決）")
		}
		fmt.Fprintln(s.writer(), puzzle.String(-1))
		return nil
	}

	// 仮定法が必要 → 複数解の可能性を探索（最大2つまで）
	var solutions []*Puzzle
	maxDepth := 0
	s.searchSolutions(puzzle, &solutions, 2, 1, &maxDepth)

	if s.Logging {
		fmt.Fprintf(s.writer(), "仮定法の深さ: %d\n", maxDepth)
	}

	switch len(solutions) {
	case 0:
		return errors.New("解がありません。")
	case 1:
		fmt.Fprintln(s.writer(), solutions[0].String(-1))
	default:
		fmt.Fprintln(s.writer(), "複数解が存在します。")
		for i, sol := range solutions {
			fmt.Fprintf(s.writer(), "--- 解 %d ---\n", i+1)
			fmt.Fprintln(s.writer(), sol.String(-1))
		}
	}
	return nil
}

// searchSolutions は再帰的にパズルを解き、解を最大maxSolutions個まで探索する
// depth: 現在の仮定の深さ、maxDepth: 解が見つかった時の最大深さを記録
func (s *Solver) searchSolutions(puzzle *Puzzle, solutions *[]*Puzzle, maxSolutions int, depth int, maxDepth *int) {
	if len(*solutions) >= maxSolutions {
		return
	}

	// ON を仮定
	onPuzzle := puzzle.Dup()
	onPuzzle.Guess(ON)
	_, err := onPuzzle.Scan()
	if err == nil {
		if onPuzzle.IsSolved() {
			*solutions = append(*solutions, onPuzzle.Dup())
			if depth > *maxDepth {
				*maxDepth = depth
			}
		} else {
			s.searchSolutions(onPuzzle, solutions, maxSolutions, depth+1, maxDepth)
		}
	}

	if len(*solutions) >= maxSolutions {
		return
	}

	// OFF を仮定
	offPuzzle := puzzle.Dup()
	offPuzzle.Guess(OFF)
	_, err = offPuzzle.Scan()
	if err == nil {
		if offPuzzle.IsSolved() {
			*solutions = append(*solutions, offPuzzle.Dup())
			if depth > *maxDepth {
				*maxDepth = depth
			}
		} else {
			s.searchSolutions(offPuzzle, solutions, maxSolutions, depth+1, maxDepth)
		}
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: picross <filename> [-v] [-o <logfile>]")
		return
	}
	filename := os.Args[1]
	logging := false
	outputFile := ""
	for i := 2; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "-v":
			logging = true
		case "-o":
			if i+1 < len(os.Args) {
				i++
				outputFile = os.Args[i]
			}
		}
	}

	solver := &Solver{Logging: logging}

	if outputFile != "" {
		// ファイルにUTF-8で直接書き出す
		of, err := os.Create(outputFile)
		if err != nil {
			fmt.Println("Error creating output file:", err)
			return
		}
		defer of.Close()
		solver.Writer = of
	}

	if err := solver.Solve(filename); err != nil {
		fmt.Fprintln(solver.writer(), "Error:", err)
	}
}
