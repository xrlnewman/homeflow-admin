package httpapi

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/xrlnewman/homeflow-admin/server/internal/app/dispatch"
	orderapp "github.com/xrlnewman/homeflow-admin/server/internal/app/order"
	"github.com/xrlnewman/homeflow-admin/server/internal/config"
	"github.com/xrlnewman/homeflow-admin/server/internal/domain"
	platformauth "github.com/xrlnewman/homeflow-admin/server/internal/platform/auth"
	"github.com/xrlnewman/homeflow-admin/server/internal/platform/cache"
	"github.com/xrlnewman/homeflow-admin/server/internal/platform/store"
)

type Dependencies struct {
	DB    *sql.DB
	Redis *cache.RedisLocker
}

type Server struct {
	cfg    config.Config
	store  *store.MemoryStore
	orders *orderapp.OrderService
	deps   Dependencies
}

func NewRouter(cfg config.Config, st *store.MemoryStore) *gin.Engine {
	return NewRouterWithDeps(cfg, st, Dependencies{})
}

func NewRouterWithDeps(cfg config.Config, st *store.MemoryStore, deps Dependencies) *gin.Engine {
	if st == nil {
		st = store.NewMemoryStore()
	}
	seed(st)
	s := &Server{cfg: cfg, store: st, orders: orderapp.NewService(st, deps.Redis), deps: deps}
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery(), traceMiddleware())
	r.GET("/healthz", s.health)
	api := r.Group("/api/v1")
	api.POST("/auth/login", s.login)
	protected := api.Group("")
	protected.Use(s.requireAuth())
	protected.POST("/auth/refresh", s.refresh)
	protected.POST("/auth/logout", s.logout)
	protected.GET("/auth/me", s.me)
	protected.POST("/orders", s.createOrder)
	protected.GET("/orders/:id", s.getOrder)
	protected.POST("/orders/:id/cancel", s.cancelOrder)
	protected.POST("/orders/:id/confirm", s.confirmOrder)
	admin := protected.Group("/admin")
	admin.Use(requireRoles("admin", "dispatcher"))
	admin.POST("/orders/:id/assign", s.assignOrder)
	admin.GET("/dispatch/recommendations", s.recommendations)
	admin.GET("/audit-logs", s.auditLogs)
	workbench := protected.Group("/workbench")
	workbench.Use(requireRoles("technician", "admin"))
	workbench.POST("/orders/:id/accept", s.accept)
	workbench.POST("/orders/:id/arrive", s.arrive)
	workbench.POST("/orders/:id/start", s.start)
	workbench.POST("/orders/:id/complete", s.complete)
	return r
}

func seed(st *store.MemoryStore) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("demo123456"), bcrypt.DefaultCost)
	st.SeedUser(store.User{ID: "user-demo", Phone: "13800000000", PasswordHash: string(hash), Name: "演示客户", Role: "customer"})
	st.SeedUser(store.User{ID: "admin-demo", Phone: "13900000000", PasswordHash: string(hash), Name: "运营管理员", Role: "admin"})
	st.SeedUser(store.User{ID: "tech-demo", Phone: "13700000000", PasswordHash: string(hash), Name: "演示师傅", Role: "technician"})
	st.SeedTechnician(store.Technician{ID: "tech-demo", Name: "演示师傅", Skills: []string{"cleaning"}, Areas: []string{"north"}, ShiftAvailable: true, Role: "technician"})
}

func traceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Trace-Id")
		if id == "" {
			id = "trace-" + uuid.NewString()
		}
		c.Set("traceId", id)
		c.Header("X-Trace-Id", id)
		c.Next()
	}
}
func (s *Server) envelope(c *gin.Context, code int, message string, data any) {
	trace, _ := c.Get("traceId")
	c.JSON(code, gin.H{"code": codeValue(code), "message": message, "data": data, "traceId": trace})
}
func codeValue(status int) any {
	if status < 400 {
		return 0
	}
	switch status {
	case http.StatusUnauthorized:
		return "AUTH_REQUIRED"
	case http.StatusForbidden:
		return "FORBIDDEN"
	case http.StatusConflict:
		return "IDEMPOTENCY_CONFLICT"
	case http.StatusBadRequest:
		return "VALIDATION_FAILED"
	default:
		return "INTERNAL_ERROR"
	}
}

func (s *Server) health(c *gin.Context) {
	data := gin.H{"status": "ok", "mysql": "not_configured", "redis": "not_configured"}
	ctx, cancel := timeLimit(c)
	defer cancel()
	status := http.StatusOK
	if s.deps.DB != nil {
		if err := s.deps.DB.PingContext(ctx); err != nil {
			data["mysql"] = "unavailable"
			status = http.StatusServiceUnavailable
		} else {
			data["mysql"] = "ok"
		}
	}
	if s.deps.Redis != nil {
		if err := s.deps.Redis.Ping(ctx); err != nil {
			data["redis"] = "unavailable"
			status = http.StatusServiceUnavailable
		} else {
			data["redis"] = "ok"
		}
	}
	s.envelope(c, status, "ok", data)
}
func timeLimit(c *gin.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.Request.Context(), 2*time.Second)
}

func (s *Server) login(c *gin.Context) {
	var in struct {
		Phone    string `json:"phone"`
		Password string `json:"password"`
	}
	if c.ShouldBindJSON(&in) != nil || in.Phone == "" || in.Password == "" {
		s.envelope(c, http.StatusBadRequest, "参数不完整", nil)
		return
	}
	u, err := s.store.UserByPhone(in.Phone)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(in.Password)) != nil {
		s.envelope(c, http.StatusUnauthorized, "账号或密码错误", nil)
		return
	}
	token, err := platformauth.Issue(s.cfg.JWTSecret, u.ID, u.Role, 2*time.Hour)
	if err != nil {
		s.envelope(c, http.StatusInternalServerError, "登录失败", nil)
		return
	}
	s.store.AddAudit(store.AuditLog{ID: uuid.NewString(), ActorID: u.ID, Action: "login", Resource: "auth", Result: "success", CreatedAt: time.Now().UTC()})
	s.envelope(c, http.StatusOK, "ok", gin.H{"accessToken": token, "tokenType": "Bearer", "expiresIn": 7200, "user": gin.H{"id": u.ID, "name": u.Name, "role": u.Role}})
}

func (s *Server) requireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := strings.TrimSpace(strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer "))
		if raw == "" {
			s.envelope(c, http.StatusUnauthorized, "请先登录", nil)
			c.Abort()
			return
		}
		claims, err := platformauth.Parse(s.cfg.JWTSecret, raw)
		if err != nil {
			s.envelope(c, http.StatusUnauthorized, "登录已失效", nil)
			c.Abort()
			return
		}
		c.Set("claims", claims)
		c.Next()
	}
}
func requireRoles(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, _ := c.Get("claims")
		role := claims.(platformauth.Claims).Role
		for _, allowed := range roles {
			if role == allowed {
				c.Next()
				return
			}
		}
		c.JSON(http.StatusForbidden, gin.H{"code": "FORBIDDEN", "message": "没有权限", "data": nil, "traceId": c.GetString("traceId")})
		c.Abort()
	}
}
func claimsOf(c *gin.Context) platformauth.Claims {
	claims, _ := c.Get("claims")
	return claims.(platformauth.Claims)
}

func (s *Server) me(c *gin.Context) {
	u, err := s.store.UserByID(claimsOf(c).UserID)
	if err != nil {
		s.envelope(c, http.StatusUnauthorized, "用户不存在", nil)
		return
	}
	s.envelope(c, http.StatusOK, "ok", gin.H{"id": u.ID, "name": u.Name, "phone": u.Phone, "role": u.Role})
}
func (s *Server) refresh(c *gin.Context) {
	cl := claimsOf(c)
	token, _ := platformauth.Issue(s.cfg.JWTSecret, cl.UserID, cl.Role, 2*time.Hour)
	s.envelope(c, http.StatusOK, "ok", gin.H{"accessToken": token, "tokenType": "Bearer", "expiresIn": 7200})
}
func (s *Server) logout(c *gin.Context) { s.envelope(c, http.StatusOK, "ok", gin.H{}) }

func (s *Server) createOrder(c *gin.Context) {
	var in orderapp.CreateInput
	if c.ShouldBindJSON(&in) != nil {
		s.envelope(c, http.StatusBadRequest, "参数不合法", nil)
		return
	}
	in.UserID = claimsOf(c).UserID
	in.IdempotencyKey = c.GetHeader("Idempotency-Key")
	order, err := s.orders.Create(c.Request.Context(), in)
	if err != nil {
		s.orderError(c, err)
		return
	}
	s.envelope(c, http.StatusCreated, "ok", order)
}
func (s *Server) getOrder(c *gin.Context) {
	order, err := s.store.OrderByID(c.Param("id"))
	if err != nil {
		s.envelope(c, http.StatusNotFound, "订单不存在", nil)
		return
	}
	cl := claimsOf(c)
	if cl.Role == "customer" && order.UserID != cl.UserID {
		s.envelope(c, http.StatusForbidden, "没有权限", nil)
		return
	}
	s.envelope(c, http.StatusOK, "ok", order)
}
func (s *Server) cancelOrder(c *gin.Context) { s.transition(c, domain.OrderCancelled) }
func (s *Server) confirmOrder(c *gin.Context) {
	s.transition(c, domain.OrderPendingCustomerConfirmation)
}
func (s *Server) transition(c *gin.Context, to domain.OrderState) {
	order, err := s.orders.Transition(c.Request.Context(), c.Param("id"), claimsOf(c).UserID, to)
	if err != nil {
		s.orderError(c, err)
		return
	}
	s.envelope(c, http.StatusOK, "ok", order)
}
func (s *Server) assignOrder(c *gin.Context) {
	var in struct {
		TechnicianID string `json:"technicianId"`
	}
	if c.ShouldBindJSON(&in) != nil || in.TechnicianID == "" {
		s.envelope(c, http.StatusBadRequest, "参数不合法", nil)
		return
	}
	order, err := s.orders.Assign(c.Request.Context(), c.Param("id"), claimsOf(c).UserID, in.TechnicianID)
	if err != nil {
		s.orderError(c, err)
		return
	}
	s.store.AddAudit(store.AuditLog{ID: uuid.NewString(), ActorID: claimsOf(c).UserID, Action: "assign", Resource: order.ID, Result: "success", CreatedAt: time.Now().UTC()})
	s.envelope(c, http.StatusOK, "ok", order)
}
func (s *Server) recommendations(c *gin.Context) {
	candidates := make([]dispatch.TechnicianCandidate, 0)
	for _, t := range s.store.Technicians() {
		candidates = append(candidates, dispatch.TechnicianCandidate{ID: t.ID, Skills: t.Skills, Areas: t.Areas, ShiftAvailable: t.ShiftAvailable, Load: t.Load})
	}
	s.envelope(c, http.StatusOK, "ok", dispatch.RankTechnicians(candidates, "cleaning", "north"))
}
func (s *Server) auditLogs(c *gin.Context) {
	s.envelope(c, http.StatusOK, "ok", gin.H{"list": s.store.Audits(), "total": len(s.store.Audits()), "page": 1, "pageSize": 20})
}
func (s *Server) accept(c *gin.Context)   { s.transition(c, domain.OrderEnRoute) }
func (s *Server) arrive(c *gin.Context)   { s.transition(c, domain.OrderServing) }
func (s *Server) start(c *gin.Context)    { s.transition(c, domain.OrderPendingCustomerConfirmation) }
func (s *Server) complete(c *gin.Context) { s.transition(c, domain.OrderCompleted) }
func (s *Server) orderError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, orderapp.ErrSlotUnavailable):
		s.envelope(c, http.StatusConflict, "预约时段已满", nil)
	case errors.Is(err, store.ErrIdempotencyConflict):
		s.envelope(c, http.StatusConflict, "幂等键已被其他请求使用", nil)
	case errors.Is(err, domain.ErrOrderStateInvalid):
		s.envelope(c, http.StatusBadRequest, "订单状态不允许该操作", nil)
	default:
		s.envelope(c, http.StatusBadRequest, err.Error(), nil)
	}
}
