package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Config struct {
	BaseURL     string
	Mode        string
	Concurrency int
	CourseID    int
	SuperID     int
	JWTSecret   string
}

type TestResult struct {
	WorkerID   int
	StudentID  int
	Username   string
	StatusCode int
	Duration   time.Duration
	Success    bool
	Message    string
	ErrorCode  string
}

type accessClaims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func generateToken(secret string, userID int, username string) string {
	now := time.Now()
	claims := accessClaims{
		Username: username,
		Role:     "user",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.Itoa(userID),
			Issuer:    "praktikum-backend",
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(1 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		panic(err)
	}
	return signed
}

func main() {
	cfg := Config{}
	flag.StringVar(&cfg.BaseURL, "base-url", "http://localhost:3000/api/v1", "Base URL backend API")
	flag.StringVar(&cfg.Mode, "mode", "supervisor", "Mode pengujian: crowd | credits | swap | supervisor")
	flag.IntVar(&cfg.Concurrency, "concurrency", 30, "Jumlah request konkuren")
	flag.IntVar(&cfg.CourseID, "course-id", 13, "ID mata kuliah untuk mode crowd (default: 13 IF399 Kelas Rebutan)")
	flag.IntVar(&cfg.SuperID, "supervisor-id", 1, "ID dosen pembimbing untuk mode supervisor (default: 1 Dr. Ir. Budi Rahardjo)")
	flag.StringVar(&cfg.JWTSecret, "jwt-secret", "supersecretjwtkeypanjangminimal32karakter!123", "JWT Secret key")
	flag.Parse()

	fmt.Println("================================================================================")
	fmt.Println("       PRAKTIKUM PEMROGRAMAN BACKEND LANJUT - MODUL 9 CONCURRENCY TESTER        ")
	fmt.Println("================================================================================")
	fmt.Printf(" Target Base URL : %s\n", cfg.BaseURL)
	fmt.Printf(" Mode Pengujian  : %s\n", cfg.Mode)
	fmt.Printf(" Konkurensi      : %d worker\n", cfg.Concurrency)
	fmt.Println("--------------------------------------------------------------------------------")

	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	// 1. Health check target server
	resp, err := client.Get(cfg.BaseURL + "/health")
	if err != nil || resp.StatusCode != http.StatusOK {
		fmt.Printf("[FATAL] Server tidak dapat dihubungi di %s/health: %v\n", cfg.BaseURL, err)
		os.Exit(1)
	}
	resp.Body.Close()
	fmt.Println("[OK] Server target sehat dan siap menerima beban.")

	switch cfg.Mode {
	case "supervisor":
		runSupervisorTest(client, cfg)
	case "crowd":
		runCrowdTest(client, cfg)
	case "credits":
		runCreditsTest(client, cfg)
	case "swap":
		runSwapTest(client, cfg)
	default:
		fmt.Printf("[ERROR] Mode tidak dikenali: %s (pilihan: crowd, credits, swap, supervisor)\n", cfg.Mode)
		os.Exit(1)
	}
}

// -----------------------------------------------------------------------------
// MODE SUPERVISOR: 30 mahasiswa mendaftar serentak ke 1 dosen (kuota 3)
// -----------------------------------------------------------------------------
func runSupervisorTest(client *http.Client, cfg Config) {
	fmt.Printf("\n[1/3] Menyiapkan %d kredensial token JWT mahasiswa...\n", cfg.Concurrency)
	type StudentClient struct {
		ID       int
		Username string
		Token    string
	}
	students := make([]StudentClient, cfg.Concurrency)
	for i := 1; i <= cfg.Concurrency; i++ {
		uname := fmt.Sprintf("mhs%03d", i)
		token := generateToken(cfg.JWTSecret, i, uname)
		students[i-1] = StudentClient{ID: i, Username: uname, Token: token}
	}
	fmt.Println("[OK] Seluruh token JWT berhasil di-generate secara deterministik.")

	fmt.Printf("\n[2/3] Melepaskan %d request serentak ke POST /supervisors/%d/assign (Pistol Gate)...\n", cfg.Concurrency, cfg.SuperID)

	var wg sync.WaitGroup
	start := make(chan struct{})
	results := make([]TestResult, cfg.Concurrency)

	for i := 0; i < cfg.Concurrency; i++ {
		wg.Add(1)
		idx := i
		s := students[idx]
		go func() {
			defer wg.Done()
			<-start // Menunggu sinyal pistol gate dibuka serentak

			reqBody, _ := json.Marshal(map[string]int{
				"student_id": s.ID,
			})

			t0 := time.Now()
			httpReq, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/supervisors/%d/assign", cfg.BaseURL, cfg.SuperID), bytes.NewReader(reqBody))
			httpReq.Header.Set("Authorization", "Bearer "+s.Token)
			httpReq.Header.Set("Content-Type", "application/json")

			resp, err := client.Do(httpReq)
			dur := time.Since(t0)

			res := TestResult{
				WorkerID:  idx + 1,
				StudentID: s.ID,
				Username:  s.Username,
				Duration:  dur,
			}

			if err != nil {
				res.StatusCode = 0
				res.Message = err.Error()
			} else {
				res.StatusCode = resp.StatusCode
				res.Success = (resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated)
				body, _ := io.ReadAll(resp.Body)
				resp.Body.Close()

				var apiResp struct {
					Message string `json:"message"`
					Code    string `json:"code"`
				}
				_ = json.Unmarshal(body, &apiResp)
				res.Message = apiResp.Message
				res.ErrorCode = apiResp.Code
			}

			results[idx] = res
		}()
	}

	// Tembak pistol gate
	startTime := time.Now()
	close(start)
	wg.Wait()
	totalElapsed := time.Since(startTime)

	fmt.Printf("[OK] Seluruh %d request selesai dalam %v.\n", cfg.Concurrency, totalElapsed)
	printReport("SUPERVISOR ASSIGNMENT CONCURRENCY TEST", results, totalElapsed, 3)
}

// -----------------------------------------------------------------------------
// MODE CROWD: 30 mahasiswa berebut kelas kuota 5 (IF399)
// -----------------------------------------------------------------------------
func runCrowdTest(client *http.Client, cfg Config) {
	fmt.Printf("\n[1/3] Menyiapkan %d kredensial token JWT mahasiswa...\n", cfg.Concurrency)
	type StudentClient struct {
		ID       int
		Username string
		Token    string
	}
	students := make([]StudentClient, cfg.Concurrency)
	for i := 1; i <= cfg.Concurrency; i++ {
		uname := fmt.Sprintf("mhs%03d", i)
		token := generateToken(cfg.JWTSecret, i, uname)
		students[i-1] = StudentClient{ID: i, Username: uname, Token: token}
	}
	fmt.Println("[OK] Seluruh token JWT berhasil disiapkan.")

	fmt.Printf("\n[2/3] Melepaskan %d request serentak ke POST /courses/%d/enroll...\n", cfg.Concurrency, cfg.CourseID)

	var wg sync.WaitGroup
	start := make(chan struct{})
	results := make([]TestResult, cfg.Concurrency)

	for i := 0; i < cfg.Concurrency; i++ {
		wg.Add(1)
		idx := i
		s := students[idx]
		go func() {
			defer wg.Done()
			<-start

			t0 := time.Now()
			httpReq, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/courses/%d/enroll", cfg.BaseURL, cfg.CourseID), bytes.NewReader([]byte("{}")))
			httpReq.Header.Set("Authorization", "Bearer "+s.Token)
			httpReq.Header.Set("Content-Type", "application/json")

			resp, err := client.Do(httpReq)
			dur := time.Since(t0)

			res := TestResult{
				WorkerID:  idx + 1,
				StudentID: s.ID,
				Username:  s.Username,
				Duration:  dur,
			}

			if err != nil {
				res.StatusCode = 0
				res.Message = err.Error()
			} else {
				res.StatusCode = resp.StatusCode
				res.Success = (resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated)
				body, _ := io.ReadAll(resp.Body)
				resp.Body.Close()

				var apiResp struct {
					Message string `json:"message"`
					Code    string `json:"code"`
				}
				_ = json.Unmarshal(body, &apiResp)
				res.Message = apiResp.Message
				res.ErrorCode = apiResp.Code
			}

			results[idx] = res
		}()
	}

	startTime := time.Now()
	close(start)
	wg.Wait()
	totalElapsed := time.Since(startTime)

	printReport("CROWD ENROLLMENT (KELAS REBUTAN) TEST", results, totalElapsed, 5)
}

// -----------------------------------------------------------------------------
// MODE CREDITS: 1 mahasiswa mendaftar serentak ke banyak matkul (batas 24 SKS)
// -----------------------------------------------------------------------------
func runCreditsTest(client *http.Client, cfg Config) {
	fmt.Println("\n[1/3] Menyiapkan token mhs001...")
	token := generateToken(cfg.JWTSecret, 1, "mhs001")

	courseIDs := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10} // 10 matkul @ 3 SKS = 30 SKS total
	count := len(courseIDs)

	fmt.Printf("\n[2/3] Melepaskan %d request pendaftaran serentak oleh 1 mahasiswa (uji batas 24 SKS)...\n", count)

	var wg sync.WaitGroup
	start := make(chan struct{})
	results := make([]TestResult, count)

	for i := 0; i < count; i++ {
		wg.Add(1)
		idx := i
		cID := courseIDs[idx]
		go func() {
			defer wg.Done()
			<-start

			t0 := time.Now()
			httpReq, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/courses/%d/enroll", cfg.BaseURL, cID), bytes.NewReader([]byte("{}")))
			httpReq.Header.Set("Authorization", "Bearer "+token)
			httpReq.Header.Set("Content-Type", "application/json")

			resp, err := client.Do(httpReq)
			dur := time.Since(t0)

			res := TestResult{
				WorkerID:  idx + 1,
				StudentID: 1,
				Username:  fmt.Sprintf("course_%d", cID),
				Duration:  dur,
			}

			if err != nil {
				res.StatusCode = 0
				res.Message = err.Error()
			} else {
				res.StatusCode = resp.StatusCode
				res.Success = (resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated)
				body, _ := io.ReadAll(resp.Body)
				resp.Body.Close()

				var apiResp struct {
					Message string `json:"message"`
					Code    string `json:"code"`
				}
				_ = json.Unmarshal(body, &apiResp)
				res.Message = apiResp.Message
				res.ErrorCode = apiResp.Code
			}

			results[idx] = res
		}()
	}

	startTime := time.Now()
	close(start)
	wg.Wait()
	totalElapsed := time.Since(startTime)

	printReport("CREDIT LIMIT (SKS CHECK) CONCURRENCY TEST", results, totalElapsed, 8)
}

// -----------------------------------------------------------------------------
// MODE SWAP: Uji pertukaran kelas serentak (A->B dan B->A) untuk Deadlock Prevention
// -----------------------------------------------------------------------------
func runSwapTest(client *http.Client, cfg Config) {
	fmt.Println("\n[1/3] Menyiapkan token mhs001 dan mhs002...")
	token1 := generateToken(cfg.JWTSecret, 1, "mhs001")
	token2 := generateToken(cfg.JWTSecret, 2, "mhs002")

	fmt.Println("\n[2/3] Menjalankan pertukaran bersilang serentak (Deadlock Prevention Order)...")
	var wg sync.WaitGroup
	start := make(chan struct{})
	results := make([]TestResult, 2)

	// Worker 1: mhs001 swap matkul 1 ke 2
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-start
		t0 := time.Now()
		reqBody, _ := json.Marshal(map[string]int{"leave_course_id": 1, "join_course_id": 2})
		httpReq, _ := http.NewRequest(http.MethodPost, cfg.BaseURL+"/courses/swap", bytes.NewReader(reqBody))
		httpReq.Header.Set("Authorization", "Bearer "+token1)
		httpReq.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(httpReq)
		dur := time.Since(t0)

		res := TestResult{WorkerID: 1, StudentID: 1, Username: "mhs001", Duration: dur}
		if err != nil {
			res.Message = err.Error()
		} else {
			res.StatusCode = resp.StatusCode
			res.Success = (resp.StatusCode == http.StatusOK)
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			res.Message = string(body)
		}
		results[0] = res
	}()

	// Worker 2: mhs002 swap matkul 2 ke 1
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-start
		t0 := time.Now()
		reqBody, _ := json.Marshal(map[string]int{"leave_course_id": 2, "join_course_id": 1})
		httpReq, _ := http.NewRequest(http.MethodPost, cfg.BaseURL+"/courses/swap", bytes.NewReader(reqBody))
		httpReq.Header.Set("Authorization", "Bearer "+token2)
		httpReq.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(httpReq)
		dur := time.Since(t0)

		res := TestResult{WorkerID: 2, StudentID: 2, Username: "mhs002", Duration: dur}
		if err != nil {
			res.Message = err.Error()
		} else {
			res.StatusCode = resp.StatusCode
			res.Success = (resp.StatusCode == http.StatusOK)
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			res.Message = string(body)
		}
		results[1] = res
	}()

	startTime := time.Now()
	close(start)
	wg.Wait()
	totalElapsed := time.Since(startTime)

	printReport("CROSS SWAP DEADLOCK PREVENTION TEST", results, totalElapsed, 2)
}

// -----------------------------------------------------------------------------
// REPORT PRINTER
// -----------------------------------------------------------------------------
func printReport(title string, results []TestResult, totalElapsed time.Duration, expectedSuccess int) {
	fmt.Println("\n================================================================================")
	fmt.Printf("                   HASIL PENGUJIAN: %s\n", title)
	fmt.Println("================================================================================")

	var successCount, conflictCount, errorCount int
	var durations []time.Duration

	for _, r := range results {
		durations = append(durations, r.Duration)
		switch {
		case r.Success:
			successCount++
		case r.StatusCode == http.StatusConflict || r.StatusCode == http.StatusBadRequest:
			conflictCount++
		default:
			errorCount++
		}
	}

	sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })
	var minDur, maxDur, p50, p95, p99 time.Duration
	if len(durations) > 0 {
		minDur = durations[0]
		maxDur = durations[len(durations)-1]
		p50 = durations[len(durations)*50/100]
		p95 = durations[len(durations)*95/100]
		p99 = durations[len(durations)*99/100]
	}

	fmt.Printf(" Total Permintaan       : %d\n", len(results))
	fmt.Printf(" Berhasil (HTTP 200/201): %d\n", successCount)
	fmt.Printf(" Ditolak  (HTTP 400/409): %d\n", conflictCount)
	fmt.Printf(" Error Server (HTTP 500): %d\n", errorCount)
	fmt.Printf(" Total Waktu Pengujian  : %v\n", totalElapsed)
	fmt.Println("--------------------------------------------------------------------------------")
	fmt.Println(" LATENSI REQUEST:")
	fmt.Printf("   Min : %v\n", minDur)
	fmt.Printf("   P50 : %v\n", p50)
	fmt.Printf("   P95 : %v\n", p95)
	fmt.Printf("   P99 : %v\n", p99)
	fmt.Printf("   Max : %v\n", maxDur)
	fmt.Println("--------------------------------------------------------------------------------")

	fmt.Println(" DETAIL SAMPEL RESPON:")
	limitShow := 10
	if len(results) < limitShow {
		limitShow = len(results)
	}
	for i := 0; i < limitShow; i++ {
		r := results[i]
		statusTag := "[SUCCESS]"
		if !r.Success {
			statusTag = "[BLOCKED]"
		}
		fmt.Printf("   Worker #%02d (%s) -> %s Status: %d | Durasi: %v | Info: %s\n",
			r.WorkerID, r.Username, statusTag, r.StatusCode, r.Duration, r.Message)
	}
	if len(results) > limitShow {
		fmt.Printf("   ... dan %d respon lainnya.\n", len(results)-limitShow)
	}

	fmt.Println("--------------------------------------------------------------------------------")
	fmt.Println(" VERIFIKASI INTEGRITAS DATA (DATA INTEGRITY CHECK):")
	if expectedSuccess > 0 {
		if successCount == expectedSuccess && errorCount == 0 {
			fmt.Printf("   [PASSED] Integritas Terjamin: Tepat %d request berhasil sesuai batas kuota.\n", expectedSuccess)
			fmt.Printf("   Tidak terjadi Race Condition atau Over-Allocation (Zero Over-Quota).\n")
		} else if successCount > expectedSuccess {
			fmt.Printf("   [FAILED] RACE CONDITION DETECTED! Berhasil %d (seharusnya maksimal %d).\n", successCount, expectedSuccess)
		} else {
			fmt.Printf("   [INFO] Berhasil %d request dari target kuota %d.\n", successCount, expectedSuccess)
		}
	}
	fmt.Println("================================================================================")
}
