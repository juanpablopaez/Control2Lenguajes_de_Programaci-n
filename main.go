package main

import (
	"context"
	"crypto/sha256"
	"encoding/csv"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

func SimularProofOfWork(blockData string, dificultad int) (string, int) {
	targetPrefix := strings.Repeat("0", dificultad)
	nonce := 0
	for {
		select {
		case <-powCancelCh:
			return "", nonce
		default:
		}
		data := fmt.Sprintf("%s%d", blockData, nonce)
		hashBytes := sha256.Sum256([]byte(data))
		hashString := fmt.Sprintf("%x", hashBytes)
		if strings.HasPrefix(hashString, targetPrefix) {
			return hashString, nonce
		}
		nonce++
		if nonce%10000 == 0 {
			select {
			case <-powCancelCh:
				return "", nonce
			default:
			}
		}
	}
}

var powCancelCh chan struct{}
var primesCancelCh chan struct{}

func EncontrarPrimos(max int) []int {
	var primes []int
	for i := 2; i < max; i++ {
		select {
		case <-primesCancelCh:
			return nil
		default:
		}
		isPrime := true
		for j := 2; j*j <= i; j++ {
			if i%j == 0 {
				isPrime = false
				break
			}
			if j%5000 == 0 {
				select {
				case <-primesCancelCh:
					return nil
				default:
				}
			}
		}
		if isPrime {
			primes = append(primes, i)
		}
	}
	return primes
}


func CalcularTrazaDeProductoDeMatrices(n int) int {
	m1 := make([][]int, n)
	m2 := make([][]int, n)
	for i := 0; i < n; i++ {
		m1[i] = make([]int, n)
		m2[i] = make([]int, n)
		for j := 0; j < n; j++ {
			m1[i][j] = rand.Intn(10)
			m2[i][j] = rand.Intn(10)
		}
	}
	trace := 0
	for i := 0; i < n; i++ {
		sum := 0
		for k := 0; k < n; k++ {
			sum += m1[i][k] * m2[k][i]
		}
		trace += sum
	}
	return trace
}

type TaskResult struct {
	id      string 
	ok      bool
	summary string        
	elapsed time.Duration 
}

func printSection(title string, lines ...string) {
	fmt.Printf("\n== %s ==\n", title)
	for _, ln := range lines {
		fmt.Println(" -", ln)
	}
}


type RunConfig struct {
	N         int
	Threshold int
	PowData   string
	PowDiff   int
	PrimeMax  int
}

func decideBranch(n, threshold int) (winner string, trace int) {
	trace = CalcularTrazaDeProductoDeMatrices(n)
	if trace >= threshold {
		return "A", trace
	}
	return "B", trace
}

func runBranchA(ctx context.Context, powData string, dificultad int) TaskResult {
	r := TaskResult{id: "A"}
	start := time.Now()
	hash, nonce := SimularProofOfWork(powData, dificultad)
	r.elapsed = time.Since(start)
	if hash != "" {
		r.ok = true
		r.summary = fmt.Sprintf("hash=%s nonce=%d", hash, nonce)
	} else {
		r.ok = false
		r.summary = fmt.Sprintf("cancelado nonce=%d", nonce)
	}
	return r
}

func runBranchB(ctx context.Context, max int) TaskResult {
	r := TaskResult{id: "B"}
	start := time.Now()
	primos := EncontrarPrimos(max)
	r.elapsed = time.Since(start)
	if primos != nil {
		r.ok = true
		r.summary = fmt.Sprintf("primos=%d", len(primos))
	} else {
		r.ok = false
		r.summary = "cancelado"
	}
	return r
}

type RunStats struct {
	total    time.Duration
	decision string
	trace    int
	A        TaskResult
	B        TaskResult
}

func speculativeRun(cfg RunConfig) RunStats {
	totalStart := time.Now()
	powCancelCh = make(chan struct{})
	primesCancelCh = make(chan struct{})

	resCh := make(chan TaskResult, 2)
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		resCh <- runBranchA(context.Background(), cfg.PowData, cfg.PowDiff)
	}()
	go func() {
		defer wg.Done()
		resCh <- runBranchB(context.Background(), cfg.PrimeMax)
	}()

	decision, trace := decideBranch(cfg.N, cfg.Threshold)

	if decision == "A" {
		close(primesCancelCh)
	} else {
		close(powCancelCh)
	}

	var r1, r2 TaskResult
	r1 = <-resCh
	r2 = <-resCh
	wg.Wait()
	close(resCh)

	var ra, rb TaskResult
	if r1.id == "A" {
		ra, rb = r1, r2
	} else {
		ra, rb = r2, r1
	}

	totalEnd := time.Now()
	total := totalEnd.Sub(totalStart)

	printSection("Especulativo",
		fmt.Sprintf("trace=%d, umbral=%d => %s", trace, cfg.Threshold, decision),
		fmt.Sprintf("A: ok=%v, t=%v, %s", ra.ok, ra.elapsed, ra.summary),
		fmt.Sprintf("B: ok=%v, t=%v, %s", rb.ok, rb.elapsed, rb.summary),
		fmt.Sprintf("Total: %v", total),
	)

	return RunStats{total: total, decision: decision, trace: trace, A: ra, B: rb}
}

func sequentialRun(cfg RunConfig) (total time.Duration, decision string, trace int, result TaskResult) {
	totalStart := time.Now()
	decision, trace = decideBranch(cfg.N, cfg.Threshold)

	if decision == "A" {
		br := TaskResult{id: "A"}
		start := time.Now()
		hash, nonce := SimularProofOfWork(cfg.PowData, cfg.PowDiff)
		br.elapsed = time.Since(start)
		br.ok = true
		br.summary = fmt.Sprintf("hash=%s nonce=%d", hash, nonce)
		result = br
	} else {
		br := TaskResult{id: "B"}
		start := time.Now()
		primos := EncontrarPrimos(cfg.PrimeMax)
		br.elapsed = time.Since(start)
		br.ok = true
		br.summary = fmt.Sprintf("primos=%d", len(primos))
		result = br
	}

	total = time.Since(totalStart)

	printSection("Secuencial",
		fmt.Sprintf("trace=%d, umbral=%d => %s", trace, cfg.Threshold, decision),
		fmt.Sprintf("Resultado: ok=%v, t=%v, %s", result.ok, result.elapsed, result.summary),
		fmt.Sprintf("Total: %v", total),
	)

	return total, decision, trace, result
}

func appendCSV(filename string, header []string, rows [][]string) error {
	fileExists := false
	if _, err := os.Stat(filename); err == nil {
		fileExists = true
	}

	f, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	if !fileExists && header != nil {
		if err := w.Write(header); err != nil {
			return err
		}
	}
	for _, r := range rows {
		if err := w.Write(r); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	n := flag.Int("n", 300, "Dimensión N de matrices NxN para decisión (más grande = más lento, O(n^3))")
	umbral := flag.Int("umbral", 1000000, "Umbral para decidir rama ganadora (trace >= umbral => A, si no B)")
	archivo := flag.String("archivo", "metricas.csv", "Archivo CSV de salida de métricas")
	powData := flag.String("powData", "bloque-demo", "Datos base para Proof-of-Work")
	powDiff := flag.Int("powDiff", 5, "Dificultad de Proof-of-Work (5-6 tarda segundos)")
	primosMax := flag.Int("primosMax", 500000, "Máximo para búsqueda de primos (más grande = más lento)")
	modo := flag.String("modo", "spec", "Modo: spec | seq | bench")
	runs := flag.Int("runs", 30, "Cantidad de repeticiones para bench")
	flag.Parse()
	ms := func(d time.Duration) string { return fmt.Sprintf("%.3f", float64(d.Microseconds())/1000.0) }

	cfg := RunConfig{
		N:         *n,
		Threshold: *umbral,
		PowData:   *powData,
		PowDiff:   *powDiff,
		PrimeMax:  *primosMax,
	}

	header := []string{
		"timestamp",
		"modo",
		"n",
		"umbral",
		"powDiff",
		"primosMax",
		"chosen",
		"trace",
		"total_ms",
		"a_ok", "a_ms",
		"b_ok", "b_ms",
	}

	mode := strings.ToLower(*modo)
	switch mode {
	case "spec":
		m := speculativeRun(cfg)
		rows := [][]string{{
			time.Now().UTC().Format(time.RFC3339),
			"spec",
			strconv.Itoa(cfg.N),
			strconv.Itoa(cfg.Threshold),
			strconv.Itoa(cfg.PowDiff),
			strconv.Itoa(cfg.PrimeMax),
			m.decision,
			strconv.Itoa(m.trace),
			ms(m.total),
			strconv.FormatBool(m.A.ok),
			ms(m.A.elapsed),
			strconv.FormatBool(m.B.ok),
			ms(m.B.elapsed),
		}}
		if err := appendCSV(*archivo, header, rows); err != nil {
			fmt.Println("Error escribiendo CSV:", err)
		}

		if m.decision == "A" && m.A.ok {
			printSection("Resumen A",
				m.A.summary,
				fmt.Sprintf("Total: %v", m.total),
			)
		} else if m.decision == "B" && m.B.ok {
			printSection("Resumen B",
				m.B.summary,
				fmt.Sprintf("Total: %v", m.total),
			)
		}

	case "seq":
		total, decision, trace, br := sequentialRun(cfg)
		rows := [][]string{{
			time.Now().UTC().Format(time.RFC3339),
			"seq",
			strconv.Itoa(cfg.N),
			strconv.Itoa(cfg.Threshold),
			strconv.Itoa(cfg.PowDiff),
			strconv.Itoa(cfg.PrimeMax),
			decision,
			strconv.Itoa(trace),
			ms(total),
			"", "", "", "",
		}}
		if err := appendCSV(*archivo, header, rows); err != nil {
			fmt.Println("Error escribiendo CSV:", err)
		}
		printSection("Secuencial (fin)",
			fmt.Sprintf("Rama %s: %s", br.id, br.summary),
			fmt.Sprintf("Total: %v", total),
		)

	case "bench":
		outRows := [][]string{}
		var sumSpec time.Duration
		var sumSeq time.Duration

		fmt.Println(">> Bench especulativo")
		for i := 0; i < *runs; i++ {
			m := speculativeRun(cfg)
			sumSpec += m.total
			outRows = append(outRows, []string{
				time.Now().UTC().Format(time.RFC3339),
				"spec",
				strconv.Itoa(cfg.N),
				strconv.Itoa(cfg.Threshold),
				strconv.Itoa(cfg.PowDiff),
				strconv.Itoa(cfg.PrimeMax),
				m.decision,
				strconv.Itoa(m.trace),
				ms(m.total),
				strconv.FormatBool(m.A.ok),
				ms(m.A.elapsed),
				strconv.FormatBool(m.B.ok),
				ms(m.B.elapsed),
			})
		}

		fmt.Println(">> Bench secuencial")
		for i := 0; i < *runs; i++ {
			total, decision, trace, _ := sequentialRun(cfg)
			sumSeq += total
			outRows = append(outRows, []string{
				time.Now().UTC().Format(time.RFC3339),
				"seq",
				strconv.Itoa(cfg.N),
				strconv.Itoa(cfg.Threshold),
				strconv.Itoa(cfg.PowDiff),
				strconv.Itoa(cfg.PrimeMax),
				decision,
				strconv.Itoa(trace),
				ms(total),
				"", "", "", "",
			})
		}

		avgSpec := time.Duration(int64(sumSpec) / int64(*runs))
		avgSeq := time.Duration(int64(sumSeq) / int64(*runs))
		speedup := float64(avgSeq.Microseconds()) / float64(avgSpec.Microseconds())

		if err := appendCSV(*archivo, header, outRows); err != nil {
			fmt.Println("Error escribiendo CSV:", err)
		}

		fmt.Println()
		printSection("Benchmark – resumen",
			fmt.Sprintf("Promedio especulativo: %v", avgSpec),
			fmt.Sprintf("Promedio secuencial:  %v", avgSeq),
			fmt.Sprintf("Speedup = Tseq/Tspec = %.3f", speedup),
		)

	default:
		fmt.Println("Modo no reconocido. Usa -modo=spec | seq | bench")
	}
}
