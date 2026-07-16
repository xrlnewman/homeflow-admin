package order

import (
	"context"
	"testing"
	"time"

	"github.com/xrlnewman/homeflow-admin/server/internal/domain"
)

func TestCreateOrderRejectsFullSlot(t *testing.T) {
	store := NewMemoryStore()
	store.SeedService(Service{ID: "svc-clean", Name: "深度保洁", SlotCapacity: 1})
	store.SeedSlot(Slot{ID: "slot-am", ServiceID: "svc-clean", Date: "2026-07-18", StartsAt: time.Date(2026, 7, 18, 9, 0, 0, 0, time.UTC), Capacity: 1, Used: 1})
	svc := NewService(store, nil)
	_, err := svc.Create(context.Background(), CreateInput{UserID: "u-1", ServiceID: "svc-clean", AddressID: "addr-1", Date: "2026-07-18", SlotID: "slot-am", IdempotencyKey: "id-1"})
	if err == nil || err != ErrSlotUnavailable {
		t.Fatalf("expected ErrSlotUnavailable, got %v", err)
	}
}

func TestCreateOrderIsIdempotent(t *testing.T) {
	store := NewMemoryStore()
	store.SeedService(Service{ID: "svc-clean", Name: "深度保洁", SlotCapacity: 1})
	store.SeedSlot(Slot{ID: "slot-am", ServiceID: "svc-clean", Date: "2026-07-18", StartsAt: time.Date(2026, 7, 18, 9, 0, 0, 0, time.UTC), Capacity: 1})
	svc := NewService(store, nil)
	in := CreateInput{UserID: "u-1", ServiceID: "svc-clean", AddressID: "addr-1", Date: "2026-07-18", SlotID: "slot-am", IdempotencyKey: "id-same"}
	first, err := svc.Create(context.Background(), in)
	if err != nil {
		t.Fatalf("first create: %v", err)
	}
	second, err := svc.Create(context.Background(), in)
	if err != nil {
		t.Fatalf("second create: %v", err)
	}
	if first.ID != second.ID || store.OrderCount() != 1 {
		t.Fatalf("expected one order, first=%s second=%s count=%d", first.ID, second.ID, store.OrderCount())
	}
	if first.State != domain.OrderPendingConfirmation {
		t.Fatalf("unexpected initial state %s", first.State)
	}
}
