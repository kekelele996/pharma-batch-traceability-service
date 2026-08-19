package notification

import (
	"testing"

	"pharma-batch-traceability-service/internal/platform"
)

func TestEnqueueNoPanic04(t *testing.T) {
	svc := NewService(platform.NewClock())
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Enqueue panicked: %v", r)
		}
	}()
	svc.Enqueue("sms", "13800000000", "subject", "body")
	if got := svc.RecipientsByChannel("sms"); len(got) != 1 || got[0] != "13800000000" {
		t.Fatalf("unexpected recipients: %v", got)
	}
}
