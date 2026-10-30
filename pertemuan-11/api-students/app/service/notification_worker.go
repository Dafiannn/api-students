package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/hibiken/asynq"

	"api-students-v11/app/model"
	"api-students-v11/app/repository"
	"api-students-v11/helper"
)

// NotificationWorker mengerjakan task "notification:send".
//
// Bagi worker, Asynq adalah framework-nya — sama seperti Fiber bagi API.
// Karena itu method di sini menerima *asynq.Task, seperti method service
// lain menerima *fiber.Ctx.
type NotificationWorker struct {
	users         repository.UserRepository
	notifications repository.NotificationRepository
	courses       repository.CourseRepository
	tx            *repository.TxManager
	mailer        helper.Mailer
	logger        *slog.Logger
}

func NewNotificationWorker(
	users repository.UserRepository,
	notifications repository.NotificationRepository,
	courses repository.CourseRepository,
	tx *repository.TxManager,
	mailer helper.Mailer,
	logger *slog.Logger,
) *NotificationWorker {
	return &NotificationWorker{
		users:         users,
		notifications: notifications,
		courses:       courses,
		tx:            tx,
		mailer:        mailer,
		logger:        logger,
	}
}

// HandleSend mengembalikan error untuk meminta Asynq melakukan retry nanti,
// atau error yang membungkus asynq.SkipRetry untuk menyatakan "jangan di-retry,
// task ini tidak mungkin berhasil".
//
// Asynq menjamin setiap task dijalankan PALING SEDIKIT satu kali, bukan
// TEPAT satu kali. Task yang sama bisa datang dua kali, misalnya bila
// worker mati setelah mengirim email tetapi sebelum sempat melapor ke Redis.
// Karena itu handler ini harus idempoten: menjalankannya dua kali untuk
// task yang sama harus berakibat sama dengan menjalankannya sekali.
// FIX PUZZLE #6 (compiler): ganti t asynq.Task menjadi t *asynq.Task
func (w *NotificationWorker) HandleSend(ctx context.Context, t *asynq.Task) error {
	var job model.NotificationJob
	if err := json.Unmarshal(t.Payload(), &job); err != nil {
		return fmt.Errorf("%w: payload task rusak: %v", asynq.SkipRetry, err)
	}

	// Tugas Mandiri D.2: Event perubahan kuota kelas
	if job.EventType == model.EventQuotaChanged {
		return w.handleQuotaChanged(ctx, job)
	}

	var event model.EnrollmentEvent
	if err := json.Unmarshal(job.Payload, &event); err != nil {
		return fmt.Errorf("%w: payload event %d rusak: %v", asynq.SkipRetry, job.OutboxID, err)
	}

	user, err := w.users.FindByID(ctx, event.UserID)
	if errors.Is(err, repository.ErrNotFound) {
		// Akunnya sudah dihapus. Melakukan retry tidak akan memunculkannya.
		return fmt.Errorf("%w: user %d tidak ditemukan", asynq.SkipRetry, event.UserID)
	}
	if err != nil {
		return err // database sedang bermasalah: layak di-retry
	}

	msg, err := BuildMessage(job.EventType, event, user)
	if err != nil {
		return fmt.Errorf("%w: %v", asynq.SkipRetry, err)
	}

	alreadySent := false
	err = w.tx.WithTxOnce(ctx, func(tx repository.DBTX) error {
		// FIX PUZZLE #8 (logic): gunakan w.notifications.WithTx(tx), bukan w.notifications langsung!
		// Claim dan pengiriman berada di dalam SATU transaction:
		// - pengiriman gagal -> claim ikut dibatalkan -> attempt
		// berikutnya masih dapat mengirim;
		// - pengiriman sukses -> claim di-commit -> attempt berikutnya
		// melihat claim ini dan tidak mengirim lagi.
		claimed, err := w.notifications.WithTx(tx).Claim(ctx, model.Notification{
			OutboxID: job.OutboxID,
			UserID:   user.ID,
			Channel:  "email",
			Subject:  msg.Subject,
			Body:     msg.Body,
		})
		if err != nil {
			return err
		}
		if !claimed {
			alreadySent = true
			return nil
		}
		return w.mailer.Send(ctx, msg)
	})

	// FIX PUZZLE #7 (logic): bungkus asynq.SkipRetry dengan %w agar errors.Is mengenalinya!
	if errors.Is(err, helper.ErrPermanentDelivery) {
		return fmt.Errorf("%w: %v", asynq.SkipRetry, err)
	}
	if err != nil {
		return err
	}

	retry, _ := asynq.GetRetryCount(ctx)
	w.logger.Info("notification_processed",
		slog.Int64("outbox_id", job.OutboxID),
		slog.String("event", job.EventType),
		slog.Int("retry", retry),
		slog.Bool("already_sent", alreadySent))
	return nil
}

func (w *NotificationWorker) handleQuotaChanged(ctx context.Context, job model.NotificationJob) error {
	var event model.QuotaChangedEvent
	if err := json.Unmarshal(job.Payload, &event); err != nil {
		return fmt.Errorf("%w: payload event quota %d rusak: %v", asynq.SkipRetry, job.OutboxID, err)
	}

	userIDs, err := w.courses.FindStudentsByCourse(ctx, event.CourseID)
	if err != nil {
		return fmt.Errorf("ambil peserta kelas: %w", err)
	}

	for _, uid := range userIDs {
		user, err := w.users.FindByID(ctx, uid)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				continue
			}
			return err
		}

		msg := BuildQuotaMessage(event, user)
		alreadySent := false

		err = w.tx.WithTxOnce(ctx, func(tx repository.DBTX) error {
			claimed, err := w.notifications.WithTx(tx).Claim(ctx, model.Notification{
				OutboxID: job.OutboxID,
				UserID:   user.ID,
				Channel:  "email",
				Subject:  msg.Subject,
				Body:     msg.Body,
			})
			if err != nil {
				return err
			}
			if !claimed {
				alreadySent = true
				return nil
			}
			return w.mailer.Send(ctx, msg)
		})

		if errors.Is(err, helper.ErrPermanentDelivery) {
			w.logger.Warn("permanent_failure_skipped", slog.Int("user_id", uid), slog.String("error", err.Error()))
			continue
		}
		if err != nil {
			return err
		}

		retry, _ := asynq.GetRetryCount(ctx)
		w.logger.Info("notification_processed",
			slog.Int64("outbox_id", job.OutboxID),
			slog.String("event", job.EventType),
			slog.Int("retry", retry),
			slog.Bool("already_sent", alreadySent))
	}
	return nil
}
