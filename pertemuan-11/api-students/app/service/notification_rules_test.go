package service

import (
	"testing"

	"api-students-v11/app/model"
)

func TestBuildMessage(t *testing.T) {
	user := model.User{
		ID:       1,
		Username: "mhsuji",
		Email:    "mhsuji@student.unair.ac.id",
	}

	event := model.EnrollmentEvent{
		UserID:         1,
		CourseID:       2,
		CourseCode:     "IF102",
		CourseName:     "Basis Data Lanjut",
		FromCourseCode: "IF101",
		FromCourseName: "Algoritma",
	}

	t.Run("EventEnrollmentCreated", func(t *testing.T) {
		msg, err := BuildMessage(model.EventEnrollmentCreated, event, user)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expectedSub := "KRS: berhasil terdaftar di IF102 Basis Data Lanjut"
		if msg.Subject != expectedSub {
			t.Errorf("expected subject %q, got %q", expectedSub, msg.Subject)
		}
		if msg.To != user.Email {
			t.Errorf("expected to %q, got %q", user.Email, msg.To)
		}
	})

	t.Run("EventEnrollmentDropped", func(t *testing.T) {
		msg, err := BuildMessage(model.EventEnrollmentDropped, event, user)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expectedSub := "KRS: keluar dari IF102 Basis Data Lanjut"
		if msg.Subject != expectedSub {
			t.Errorf("expected subject %q, got %q", expectedSub, msg.Subject)
		}
	})

	t.Run("EventEnrollmentSwapped", func(t *testing.T) {
		msg, err := BuildMessage(model.EventEnrollmentSwapped, event, user)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expectedSub := "KRS: pindah dari IF101 Algoritma ke IF102 Basis Data Lanjut"
		if msg.Subject != expectedSub {
			t.Errorf("expected subject %q, got %q", expectedSub, msg.Subject)
		}
	})

	t.Run("UnknownEvent", func(t *testing.T) {
		_, err := BuildMessage("unknown.event", event, user)
		if err == nil {
			t.Fatal("expected error for unknown event, got nil")
		}
	})
}

func TestBuildQuotaMessage(t *testing.T) {
	user := model.User{
		ID:       1,
		Username: "mhsuji",
		Email:    "mhsuji@student.unair.ac.id",
	}

	event := model.QuotaChangedEvent{
		CourseID:   10,
		CourseCode: "IF308",
		CourseName: "Visi Komputer",
		NewQuota:   50,
	}

	msg := BuildQuotaMessage(event, user)
	expectedSub := "KRS: kuota IF308 Visi Komputer berubah menjadi 50"
	if msg.Subject != expectedSub {
		t.Errorf("expected subject %q, got %q", expectedSub, msg.Subject)
	}
	if msg.To != user.Email {
		t.Errorf("expected to %q, got %q", user.Email, msg.To)
	}
}
