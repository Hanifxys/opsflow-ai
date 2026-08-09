package integration_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	aiApp "github.com/opsflow/ai-gateway/internal/application"
	"github.com/opsflow/ai-gateway/internal/domain"
	authApp "github.com/opsflow/auth-service/internal/application"
	authDomain "github.com/opsflow/auth-service/internal/domain"
	"github.com/opsflow/common/jwt"
	"github.com/opsflow/common/telemetry"
	incApp "github.com/opsflow/incident-service/internal/application"
	incDomain "github.com/opsflow/incident-service/internal/domain"
	notifApp "github.com/opsflow/notification-service/internal/application"
	notifDomain "github.com/opsflow/notification-service/internal/domain"
	regApp "github.com/opsflow/registry-service/internal/application"
	regDomain "github.com/opsflow/registry-service/internal/domain"
)

// In-Memory Repositories for E2E integration validation

type mockAuthRepo struct {
	users map[string]*authDomain.User
}

func (m *mockAuthRepo) CreateUser(ctx context.Context, u *authDomain.User) error {
	m.users[u.Email] = u
	return nil
}
func (m *mockAuthRepo) GetUserByEmail(ctx context.Context, email string) (*authDomain.User, error) {
	u, ok := m.users[email]
	if !ok {
		return nil, authDomain.ErrUserNotFound
	}
	return u, nil
}
func (m *mockAuthRepo) GetUserByID(ctx context.Context, id uuid.UUID) (*authDomain.User, error) {
	for _, u := range m.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, authDomain.ErrUserNotFound
}

type mockHasher struct{}

func (m *mockHasher) HashPassword(password string) (string, error) { return "hashed_" + password, nil }
func (m *mockHasher) ComparePassword(hashedPassword, password string) error {
	if hashedPassword != "hashed_"+password {
		return authDomain.ErrInvalidCredentials
	}
	return nil
}

type mockServiceRepo struct {
	services map[uuid.UUID]*regDomain.Service
}

func (m *mockServiceRepo) Create(ctx context.Context, s *regDomain.Service) error {
	m.services[s.ID] = s
	return nil
}
func (m *mockServiceRepo) GetByID(ctx context.Context, id uuid.UUID) (*regDomain.Service, error) {
	s, ok := m.services[id]
	if !ok {
		return nil, regDomain.ErrServiceNotFound
	}
	return s, nil
}
func (m *mockServiceRepo) List(ctx context.Context, limit, offset int) ([]*regDomain.Service, error) {
	var res []*regDomain.Service
	for _, s := range m.services {
		res = append(res, s)
	}
	return res, nil
}
func (m *mockServiceRepo) Update(ctx context.Context, s *regDomain.Service) error {
	m.services[s.ID] = s
	return nil
}
func (m *mockServiceRepo) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.services, id)
	return nil
}
func (m *mockServiceRepo) AddEnvironment(ctx context.Context, env *regDomain.ServiceEnvironment) error {
	return nil
}
func (m *mockServiceRepo) ListEnvironments(ctx context.Context, serviceID uuid.UUID) ([]*regDomain.ServiceEnvironment, error) {
	return nil, nil
}
func (m *mockServiceRepo) AddDependency(ctx context.Context, dep *regDomain.ServiceDependency) error {
	return nil
}
func (m *mockServiceRepo) ListDependencies(ctx context.Context, serviceID uuid.UUID) ([]*regDomain.ServiceDependency, error) {
	return nil, nil
}

type mockIncidentRepo struct {
	incidents map[uuid.UUID]*incDomain.Incident
}

func (m *mockIncidentRepo) Create(ctx context.Context, inc *incDomain.Incident) error {
	m.incidents[inc.ID] = inc
	return nil
}
func (m *mockIncidentRepo) GetByID(ctx context.Context, id uuid.UUID) (*incDomain.Incident, error) {
	inc, ok := m.incidents[id]
	if !ok {
		return nil, incDomain.ErrIncidentNotFound
	}
	return inc, nil
}
func (m *mockIncidentRepo) List(ctx context.Context, filter incDomain.IncidentFilter) ([]*incDomain.Incident, error) {
	var res []*incDomain.Incident
	for _, inc := range m.incidents {
		res = append(res, inc)
	}
	return res, nil
}
func (m *mockIncidentRepo) Update(ctx context.Context, inc *incDomain.Incident) error {
	m.incidents[inc.ID] = inc
	return nil
}
func (m *mockIncidentRepo) AddTimelineEvent(ctx context.Context, event *incDomain.TimelineEvent) error {
	return nil
}
func (m *mockIncidentRepo) AddComment(ctx context.Context, comment *incDomain.IncidentComment) error {
	return nil
}

type mockNotifRepo struct {
	notifications map[string]*notifDomain.Notification
}

func (m *mockNotifRepo) SaveIfNew(ctx context.Context, n *notifDomain.Notification) (bool, error) {
	if _, exists := m.notifications[n.IdempotencyKey]; exists {
		return false, nil
	}
	m.notifications[n.IdempotencyKey] = n
	return true, nil
}
func (m *mockNotifRepo) ListRecent(ctx context.Context, limit int) ([]*notifDomain.Notification, error) {
	var res []*notifDomain.Notification
	for _, n := range m.notifications {
		res = append(res, n)
	}
	return res, nil
}

type mockAIRepo struct {
	conversations map[uuid.UUID]*domain.Conversation
	messages      map[uuid.UUID][]domain.Message
	approvals     map[uuid.UUID]*domain.ApprovalRequest
}

func (m *mockAIRepo) CreateConversation(ctx context.Context, conv *domain.Conversation) error {
	m.conversations[conv.ID] = conv
	return nil
}
func (m *mockAIRepo) GetConversationByID(ctx context.Context, id uuid.UUID) (*domain.Conversation, error) {
	c, ok := m.conversations[id]
	if !ok {
		return nil, domain.ErrConversationNotFound
	}
	return c, nil
}
func (m *mockAIRepo) ListConversationsByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Conversation, error) {
	var list []*domain.Conversation
	for _, c := range m.conversations {
		if c.UserID == userID {
			list = append(list, c)
		}
	}
	return list, nil
}
func (m *mockAIRepo) SaveMessage(ctx context.Context, msg *domain.Message) error {
	m.messages[msg.ConversationID] = append(m.messages[msg.ConversationID], *msg)
	return nil
}
func (m *mockAIRepo) ListMessagesByConversationID(ctx context.Context, convID uuid.UUID) ([]domain.Message, error) {
	return m.messages[convID], nil
}
func (m *mockAIRepo) CreateApprovalRequest(ctx context.Context, app *domain.ApprovalRequest) error {
	m.approvals[app.ID] = app
	return nil
}
func (m *mockAIRepo) GetApprovalRequestByID(ctx context.Context, id uuid.UUID) (*domain.ApprovalRequest, error) {
	app, ok := m.approvals[id]
	if !ok {
		return nil, domain.ErrApprovalNotFound
	}
	return app, nil
}
func (m *mockAIRepo) UpdateApprovalStatus(ctx context.Context, id uuid.UUID, status domain.ApprovalStatus, approvedBy uuid.UUID) error {
	app, ok := m.approvals[id]
	if !ok {
		return domain.ErrApprovalNotFound
	}
	app.Status = status
	app.ApprovedBy = &approvedBy
	return nil
}
func (m *mockAIRepo) ListPendingApprovals(ctx context.Context) ([]*domain.ApprovalRequest, error) {
	var list []*domain.ApprovalRequest
	for _, a := range m.approvals {
		if a.Status == domain.ApprovalStatusPending {
			list = append(list, a)
		}
	}
	return list, nil
}
func (m *mockAIRepo) SaveKnowledgeDocument(ctx context.Context, doc *domain.KnowledgeDocument) error {
	return nil
}
func (m *mockAIRepo) SearchKnowledge(ctx context.Context, query string, limit int) ([]*domain.KnowledgeDocument, error) {
	return nil, nil
}

func TestOpsFlow_EndToEndLifecycle(t *testing.T) {
	ctx := context.Background()
	jwtSecret := "test-secret"

	// 1. Auth Service — User Login
	authRepo := &mockAuthRepo{users: make(map[string]*authDomain.User)}
	hasher := &mockHasher{}
	authSvc := authApp.NewAuthService(authRepo, hasher, jwtSecret, 0, 0)

	user, err := authSvc.RegisterUser(ctx, "operator@opsflow.local", "password123", "Ops Operator", authDomain.RoleOperator)
	if err != nil {
		t.Fatalf("failed to register user: %v", err)
	}

	tokenResp, err := authSvc.Login(ctx, "operator@opsflow.local", "password123")
	if err != nil {
		t.Fatalf("failed to login user: %v", err)
	}

	claims, err := jwt.ValidateToken(tokenResp.AccessToken, jwtSecret)
	if err != nil {
		t.Fatalf("failed to validate JWT token: %v", err)
	}
	if claims.Email != "operator@opsflow.local" {
		t.Errorf("expected token email operator@opsflow.local, got %s", claims.Email)
	}

	// 2. Service Registry — Register Service
	regRepo := &mockServiceRepo{services: make(map[uuid.UUID]*regDomain.Service)}
	regSvc := regApp.NewServiceRegistryService(regRepo)
	svc, err := regSvc.RegisterService(ctx, "payment-service", "Core Payment Microservice", "Core Banking", "git@github.com:opsflow/payment.git")
	if err != nil {
		t.Fatalf("failed to register service: %v", err)
	}

	// 3. Incident Service — Declare Incident
	incRepo := &mockIncidentRepo{incidents: make(map[uuid.UUID]*incDomain.Incident)}
	incSvc := incApp.NewIncidentService(incRepo)
	inc, err := incSvc.CreateIncident(ctx, svc.ID, "Payment Database Timeout", "Database connection pool exhausted", incDomain.SeverityCritical, user.ID)
	if err != nil {
		t.Fatalf("failed to create incident: %v", err)
	}
	if inc.IncidentKey == "" {
		t.Error("expected non-empty incident key")
	}

	// 4. Notification Worker — Send Async Notification (Idempotent)
	notifRepo := &mockNotifRepo{notifications: make(map[string]*notifDomain.Notification)}
	notifSvc := notifApp.NewNotificationService(notifRepo)
	isNew, err := notifSvc.ProcessEvent(ctx, "EMAIL", "incident.created", []byte(`{"incident_id":"123","title":"Payment Timeout"}`))
	if err != nil || !isNew {
		t.Errorf("expected new notification delivery, got isNew=%v, err=%v", isNew, err)
	}

	// Duplicate delivery test
	isNewDup, err := notifSvc.ProcessEvent(ctx, "EMAIL", "incident.created", []byte(`{"incident_id":"123","title":"Payment Timeout"}`))
	if err != nil || isNewDup {
		t.Errorf("expected duplicate delivery to be ignored (idempotent), got isNew=%v", isNewDup)
	}

	// 5. AI Gateway — Human-in-the-Loop Incident Mitigation
	aiRepo := &mockAIRepo{
		conversations: make(map[uuid.UUID]*domain.Conversation),
		messages:      make(map[uuid.UUID][]domain.Message),
		approvals:     make(map[uuid.UUID]*domain.ApprovalRequest),
	}
	aiSvc := aiApp.NewAIService(aiRepo, nil, nil) // Will test approval workflow state
	conv, err := aiSvc.CreateConversation(ctx, user.ID, "Incident Mitigation Session")
	if err != nil {
		t.Fatalf("failed to create AI conversation: %v", err)
	}

	// Create pending approval request manually
	appReq := &domain.ApprovalRequest{
		ID:             uuid.New(),
		ConversationID: conv.ID,
		ToolName:       "restart_service",
		Arguments:      []byte(`{"service_name":"payment-service"}`),
		Status:         domain.ApprovalStatusPending,
		RequestedBy:    user.ID,
	}
	_ = aiRepo.CreateApprovalRequest(ctx, appReq)

	pendingList, err := aiSvc.ListPendingApprovals(ctx)
	if err != nil || len(pendingList) != 1 {
		t.Fatalf("expected 1 pending approval request, got %d", len(pendingList))
	}

	// Human Approves Action
	_, err = aiSvc.ApproveToolAction(ctx, appReq.ID, user.ID)
	if err != nil {
		t.Fatalf("failed to approve tool action: %v", err)
	}

	appApproved, _ := aiRepo.GetApprovalRequestByID(ctx, appReq.ID)
	if appApproved.Status != domain.ApprovalStatusApproved {
		t.Errorf("expected approval status APPROVED, got %s", appApproved.Status)
	}

	// 6. Observability — W3C Trace Context Propagation
	traceMiddleware := telemetry.TelemetryMiddleware("integration-test")(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			traceID := telemetry.TraceID(r.Context())
			if traceID == "" {
				t.Error("expected traceID in context")
			}
			w.WriteHeader(http.StatusOK)
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/incidents", nil)
	w := httptest.NewRecorder()
	traceMiddleware.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", w.Code)
	}
}
