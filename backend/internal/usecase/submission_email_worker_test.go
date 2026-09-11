package usecase

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"codeberg.org/Thr0TT1e/emh/backend/internal/config"
	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	"codeberg.org/Thr0TT1e/emh/backend/internal/smtp"
)

// mockSMTPSenderForEmails мок smtp.Sender для тестов.
type mockSMTPSenderForEmails struct {
	mu     sync.Mutex // 🆕 защита от data race при параллельной отправке
	sendFn func(ctx context.Context, msg smtp.Message) error
	calls  []smtp.Message
}

func (m *mockSMTPSenderForEmails) Send(ctx context.Context, msg smtp.Message) error {
	m.mu.Lock()
	m.calls = append(m.calls, msg)
	m.mu.Unlock()
	if m.sendFn != nil {
		return m.sendFn(ctx, msg)
	}
	return nil
}

// callsLen потокобезопасно возвращает количество отправленных писем.
func (m *mockSMTPSenderForEmails) callsLen() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.calls)
}

// findBySubject ищет письмо по теме (потокобезопасно).
// Возвращает nil если не найдено.
func (m *mockSMTPSenderForEmails) findBySubject(subject string) *smtp.Message {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := range m.calls {
		if m.calls[i].Subject == subject {
			return &m.calls[i]
		}
	}
	return nil
}

func TestSubmissionEmailWorker_EnqueueAndSend(t *testing.T) {
	sender := &mockSMTPSenderForEmails{}
	cfg := config.SubmissionEmailsConfig{
		Enabled:         true,
		ApproveSubject:  "Одобрено",
		RejectSubject:   "Отклонено",
		ApproveTemplate: "Привет, {{.SubmitterName}}! Заявка {{.SubmissionID}} одобрена.",
		RejectTemplate:  "Привет, {{.SubmitterName}}! Заявка {{.SubmissionID}} отклонена. Причина: {{.ModeratorComment}}",
		QueueSize:       10,
	}
	worker := NewSubmissionEmailWorker(sender, discardTestLogger(), cfg)

	// Запускаем воркер с не-отменяемым контекстом —
	// Run будет стабильно работать, пока Stop() не закроет канал.
	go worker.Run(context.Background())

	// Enqueue approve
	worker.Enqueue(domain.SubmissionEmailNotification{
		SubmitterName: "Иван",
		Email:         "ivan@example.com",
		SubmissionID:  "sub-001",
		Decision:      domain.ReviewDecisionApprove,
		Timestamp:     time.Now(),
	})

	// Enqueue reject
	worker.Enqueue(domain.SubmissionEmailNotification{
		SubmitterName:    "Пётр",
		Email:            "petr@example.com",
		SubmissionID:     "sub-002",
		Decision:         domain.ReviewDecisionReject,
		ModeratorComment: "Недостаточно данных",
		Timestamp:        time.Now(),
	})

	// Polling: ждём обработки с таймаутом (вместо flaky sleep или race-вызова cancel).
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if sender.callsLen() >= 2 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	// Stop() закроет канал — Run() выйдет через ok=false в select.
	// wg.Wait() дождётся завершения всех in-flight отправок.
	worker.Stop()

	// Проверяем количество
	if got := sender.callsLen(); got != 2 {
		t.Fatalf("expected 2 Send calls, got %d", got)
	}

	// Поиск по subject (порядок недетерминирован из-за параллельных горутин)
	approveMsg := sender.findBySubject("[EMH] Одобрено")
	if approveMsg == nil {
		t.Fatal("approve email not found")
	}
	if !strings.Contains(approveMsg.Body, "Иван") {
		t.Errorf("approve body should contain 'Иван', got: %s", approveMsg.Body)
	}
	if !strings.Contains(approveMsg.Body, "sub-001") {
		t.Errorf("approve body should contain 'sub-001', got: %s", approveMsg.Body)
	}

	rejectMsg := sender.findBySubject("[EMH] Отклонено")
	if rejectMsg == nil {
		t.Fatal("reject email not found")
	}
	if !strings.Contains(rejectMsg.Body, "Недостаточно данных") {
		t.Errorf("reject body should contain moderator comment, got: %s", rejectMsg.Body)
	}
	if !strings.Contains(rejectMsg.Body, "Пётр") {
		t.Errorf("reject body should contain 'Пётр', got: %s", rejectMsg.Body)
	}
}
