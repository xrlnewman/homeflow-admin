package database

import (
	"database/sql"
	"fmt"

	"github.com/xrlnewman/homeflow-admin/server/internal/domain"
	"github.com/xrlnewman/homeflow-admin/server/internal/platform/store"
)

// SQLPersistence stores order facts and audit events in MySQL 8.4.
type SQLPersistence struct{ db *sql.DB }

func NewSQLPersistence(db *sql.DB) *SQLPersistence { return &SQLPersistence{db: db} }

func (p *SQLPersistence) PersistOrder(order domain.Order, idempotencyKey string) error {
	if p == nil || p.db == nil {
		return fmt.Errorf("mysql persistence is not configured")
	}
	_, err := p.db.Exec(`INSERT INTO orders (id,user_id,service_id,address_id,service_date,slot_id,technician_id,remark,state,idempotency_key,created_at,updated_at)
VALUES (?,?,?,?,?,?,?,?,?,?,?,?)
ON DUPLICATE KEY UPDATE technician_id=VALUES(technician_id), remark=VALUES(remark), state=VALUES(state), updated_at=VALUES(updated_at)`,
		order.ID, order.UserID, order.ServiceID, order.AddressID, order.Date, order.SlotID, nullable(order.TechnicianID), nullable(order.Remark), order.State, nullable(idempotencyKey), order.CreatedAt, order.UpdatedAt)
	return err
}

func (p *SQLPersistence) PersistOrderEvent(event domain.OrderEvent) error {
	if p == nil || p.db == nil {
		return fmt.Errorf("mysql persistence is not configured")
	}
	_, err := p.db.Exec(`INSERT INTO order_events (id,order_id,from_state,to_state,actor_id,created_at) VALUES (?,?,?,?,?,?)`, event.ID, event.OrderID, nullable(string(event.From)), string(event.To), event.ActorID, event.CreatedAt)
	return err
}

func (p *SQLPersistence) PersistAudit(log store.AuditLog) error {
	if p == nil || p.db == nil {
		return fmt.Errorf("mysql persistence is not configured")
	}
	_, err := p.db.Exec(`INSERT INTO audit_logs (id,actor_id,action,resource,result,created_at) VALUES (?,?,?,?,?,?)`, log.ID, log.ActorID, log.Action, log.Resource, log.Result, log.CreatedAt)
	return err
}

func (p *SQLPersistence) PersistReview(review store.Review) error {
	if p == nil || p.db == nil {
		return fmt.Errorf("mysql persistence is not configured")
	}
	_, err := p.db.Exec(`INSERT INTO reviews (id,order_id,user_id,rating,content,created_at) VALUES (?,?,?,?,?,?)`, review.ID, review.OrderID, review.UserID, review.Rating, review.Content, review.CreatedAt)
	return err
}

func (p *SQLPersistence) PersistProof(proof store.Proof) error {
	if p == nil || p.db == nil {
		return fmt.Errorf("mysql persistence is not configured")
	}
	_, err := p.db.Exec(`INSERT INTO work_proofs (id,order_id,kind,filename,note,created_at) VALUES (?,?,?,?,?,?)`, proof.ID, proof.OrderID, proof.Kind, proof.Filename, nullable(proof.Note), proof.CreatedAt)
	return err
}

func nullable(value string) any {
	if value == "" {
		return nil
	}
	return value
}
