package store

import (
	"testing"

	"github.com/xrlnewman/homeflow-admin/server/internal/domain"
)

type recordingPersistence struct{ orders, events, audits int }

func (p *recordingPersistence) PersistOrder(domain.Order, string) error   { p.orders++; return nil }
func (p *recordingPersistence) PersistOrderEvent(domain.OrderEvent) error { p.events++; return nil }
func (p *recordingPersistence) PersistAudit(AuditLog) error               { p.audits++; return nil }

func TestMemoryStoreForwardsWritesToPersistence(t *testing.T) {
	st := NewMemoryStore()
	recorder := &recordingPersistence{}
	st.SetPersistence(recorder)
	order := domain.Order{ID: "order-1", UserID: "user-1", State: domain.OrderPendingConfirmation}
	st.SaveOrder(order, "idem-1")
	st.UpdateOrder(order, domain.OrderEvent{ID: "event-1", OrderID: order.ID, To: order.State})
	st.AddAudit(AuditLog{ID: "audit-1", ActorID: "user-1", Action: "test", Resource: order.ID, Result: "success"})
	if recorder.orders != 2 || recorder.events != 2 || recorder.audits != 1 {
		t.Fatalf("persistence writes: orders=%d events=%d audits=%d", recorder.orders, recorder.events, recorder.audits)
	}
}
