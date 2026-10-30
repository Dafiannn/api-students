package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/hibiken/asynq"

	"api-students-v11/app/model"
	"api-students-v11/config"
	"api-students-v11/database"
)

func main() {
	outboxID := flag.Int64("outbox", 0, "ID outbox yang akan di-redeliver")
	copies := flag.Int("copies", 1, "Jumlah salinan task yang dikirim ke Redis")
	flag.Parse()

	if *outboxID <= 0 {
		log.Fatal("flag -outbox wajib diisi dengan ID positif")
	}
	if *copies <= 0 {
		*copies = 1
	}

	config.LoadEnv()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := database.NewPool(ctx)
	if err != nil {
		log.Fatalf("gagal koneksi database: %v", err)
	}
	defer pool.Close()

	var eventType string
	var payload []byte
	err = pool.QueryRow(ctx, "SELECT event_type, payload FROM outbox WHERE id = $1", *outboxID).Scan(&eventType, &payload)
	if err != nil {
		log.Fatalf("gagal membaca outbox id %d: %v", *outboxID, err)
	}

	job := model.NotificationJob{
		OutboxID:  *outboxID,
		EventType: eventType,
		Payload:   payload,
	}

	jobBytes, err := json.Marshal(job)
	if err != nil {
		log.Fatalf("marshal job: %v", err)
	}

	client := asynq.NewClient(config.RedisOpt())
	defer client.Close()

	fmt.Printf("Mengirim %d salinan task untuk outbox ID %d ke queue '%s'...\n", *copies, *outboxID, model.QueueNotifications)

	for i := 1; i <= *copies; i++ {
		// Task dikirim tanpa TaskID tunggal agar Redis menerima semua copies untuk menguji idempotency worker
		task := asynq.NewTask(model.TaskSendNotification, jobBytes,
			asynq.Queue(model.QueueNotifications),
			asynq.MaxRetry(5),
			asynq.Timeout(30*time.Second),
			asynq.Retention(24*time.Hour),
		)
		info, err := client.EnqueueContext(ctx, task)
		if err != nil {
			log.Fatalf("gagal enqueue copy %d: %v", i, err)
		}
		fmt.Printf("  [copy %d] Enqueued task ID: %s\n", i, info.ID)
	}

	fmt.Println("Selesai redeliver.")
}
