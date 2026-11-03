package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf"
	"github.com/linhhuynhcoding/jss-microservices/order-service/config"
	"github.com/linhhuynhcoding/jss-microservices/order-service/internal/adapter"
	"github.com/linhhuynhcoding/jss-microservices/order-service/internal/domain"
	"github.com/linhhuynhcoding/jss-microservices/order-service/internal/repository"

	mq "github.com/linhhuynhcoding/jss-microservices/mq"
	mqconfig "github.com/linhhuynhcoding/jss-microservices/mq/config"
	"github.com/linhhuynhcoding/jss-microservices/mq/consts"

	_ "github.com/linhhuynhcoding/jss-microservices/rpc/gen/auth"
	loyaltypb "github.com/linhhuynhcoding/jss-microservices/rpc/gen/loyalty"
	notificationpb "github.com/linhhuynhcoding/jss-microservices/rpc/gen/notification"
	orderpb "github.com/linhhuynhcoding/jss-microservices/rpc/gen/order"
	productpb "github.com/linhhuynhcoding/jss-microservices/rpc/gen/product"

	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Service struct {
	orderpb.UnimplementedOrderServiceServer
	repo          *repository.OrderRepository
	authClient    *adapter.AuthClient
	productClient *adapter.ProductClient
	loyaltyClient *adapter.LoyaltyClient
	publisher     *mq.Publisher
	logger        *zap.Logger
}

func New(cfg config.Config, db *mongo.Database, log *zap.Logger) (*Service, error) {
	repo := repository.New(db)

	authClient, err := adapter.NewAuthClient(cfg.AuthServiceAddr, log)
	if err != nil {
		return nil, fmt.Errorf("failed to create auth client: %w", err)
	}
	productClient, err := adapter.NewProductClient(cfg.ProductServiceAddr, log)
	if err != nil {
		return nil, fmt.Errorf("failed to create product client: %w", err)
	}
	loyaltyClient, err := adapter.NewLoyaltyClient(cfg.LoyaltyServiceAddr, log)
	if err != nil {
		return nil, fmt.Errorf("failed to create loyalty client: %w", err)
	}

	pubCfg := mqconfig.RabbitMQConfig{
		ConnStr:       cfg.RabbitMQURL,
		ExchangeName:  consts.EXCHANGE_ORDER_SERVICE,
		ExchangeType:  "topic",
		PublisherName: consts.EXCHANGE_ORDER_SERVICE,
	}
	publisher, err := mq.NewPublisher(pubCfg, log)
	if err != nil {
		return nil, fmt.Errorf("failed to create mq publisher: %w", err)
	}

	return &Service{
		repo:          repo,
		authClient:    authClient,
		productClient: productClient,
		loyaltyClient: loyaltyClient,
		publisher:     publisher,
		logger:        log,
	}, nil
}

func (s *Service) Close() {
	if s.authClient != nil {
		s.authClient.Close()
	}
	if s.productClient != nil {
		s.productClient.Close()
	}
	if s.loyaltyClient != nil {
		s.loyaltyClient.Close()
	}
	if s.publisher != nil {
		s.publisher.Close()
	}
}

func (s *Service) CreateOrder(ctx context.Context, req *orderpb.CreateOrderRequest) (*orderpb.CreateOrderResponse, error) {
	s.logger.Info("CreateOrder called", zap.Any("req", req))

	// 1) Auth
	accessToken, err := bearerFromMD(ctx)
	if err != nil {
		return nil, err
	}
	valid, userID, role, err := s.authClient.Validate(ctx, accessToken)
	if err != nil || !valid {
		return nil, fmt.Errorf("unauthorised")
	}
	if role != "STAFF" && role != "MANAGER" && role != "ADMIN" {
		return nil, fmt.Errorf("unauthorised role: %s", role)
	}

	// 2) Validate input
	if len(req.Items) == 0 {
		return nil, fmt.Errorf("order must contain at least one product")
	}

	// 3) Next order id
	orderID, err := s.repo.NextOrderID(ctx)
	if err != nil {
		s.logger.Error("failed to get next order id", zap.Error(err))
		return nil, fmt.Errorf("failed to generate order id: %w", err)
	}

	// 4) Gọi product-service để trừ kho & lấy snapshot
	pReq := &productpb.PurchaseProductRequest{
		CustomerId: req.GetCustomerId(), // có customer (đã seed)
		OrderId:    orderID,
		Products:   make([]*productpb.PurchaseProductRequest_Product, 0, len(req.Items)),
	}
	for _, it := range req.Items {
		pReq.Products = append(pReq.Products, &productpb.PurchaseProductRequest_Product{
			ProductId: it.ProductId,
			Quantity:  it.Quantity,
		})
	}
	pResp, err := s.productClient.PurchaseProduct(ctx, pReq)
	if err != nil {
		s.logger.Error("PurchaseProduct RPC failed", zap.Error(err))
		return nil, fmt.Errorf("failed to purchase products: %w", err)
	}

	// 5) Build snapshot: price + name + image
	type snap struct {
		price float64
		name  string
		image string
	}
	byID := make(map[int32]snap, len(pResp.GetProducts()))
	for _, p := range pResp.GetProducts() {
		byID[p.GetId()] = snap{
			price: p.GetSellingPrice(),
			name:  p.GetName(),  // từ product proto
			image: p.GetImage(), // từ product proto
		}
	}

	// 6) Tính items + subtotal
	var subtotal float64
	items := make([]domain.OrderItem, 0, len(req.Items))
	for _, it := range req.Items {
		snap, ok := byID[it.ProductId]
		if !ok {
			return nil, fmt.Errorf("product snapshot not returned for product_id %d", it.ProductId)
		}
		line := snap.price * float64(it.Quantity)
		subtotal += line
		items = append(items, domain.OrderItem{
			ProductID:    it.ProductId,
			Quantity:     it.Quantity,
			UnitPrice:    snap.price,
			ProductName:  snap.name,
			ProductImage: snap.image,
			LineTotal:    line,
		})
	}

	// 7) Voucher + shipping
	shipping := req.GetShippingCost()
	discount := 0.0
	if len(req.VoucherCodes) > 0 {
		vReq := &loyaltypb.UsingVoucherRequest{
			Vouchers:            req.VoucherCodes,
			TotalProductAmount:  subtotal,
			TotalShippingAmount: shipping,
			CustomerId:          req.GetCustomerId(),
			OrderId:             orderID,
		}
		vResp, err := s.loyaltyClient.UsingVoucher(ctx, vReq)
		if err != nil {
			s.logger.Error("UsingVoucher RPC failed", zap.Error(err))
			return nil, fmt.Errorf("failed to apply vouchers: %w", err)
		}
		discount = vResp.GetTotalDiscountAmount()
	}

	final := subtotal + shipping - discount
	now := time.Now()

	// 8) Persist
	order := &domain.Order{
		OrderID:        orderID,
		CustomerName:   req.GetCustomerName(),
		CustomerID:     req.GetCustomerId(),
		StaffID:        userID,
		Items:          items,
		VoucherCodes:   req.VoucherCodes,
		TotalPrice:     subtotal,
		DiscountAmount: discount,
		FinalPrice:     final,
		ShippingCost:   shipping,
		CreatedAt:      now,
		Status:         domain.OrderStatusPending,
		StatusHistory: []domain.StatusHistory{
			{Status: domain.OrderStatusPending, Note: "created", At: now},
		},
	}
	if err := s.repo.Create(ctx, order); err != nil {
		s.logger.Error("failed to persist order", zap.Error(err))
		return nil, fmt.Errorf("failed to save order: %w", err)
	}

	// 9) Notification (không chặn flow nếu lỗi)
	evt := &notificationpb.CreateNotificationRequest{
		UserId:  order.StaffID,
		Role:    role,
		Title:   "Order created",
		Message: fmt.Sprintf("Order #%d has been created successfully", orderID),
	}
	if err := s.publisher.SendMessage(evt, "notification.create"); err != nil {
		s.logger.Error("failed to publish notification", zap.Error(err))
	}

	// 10) Publish ORDER_CREATED
	if err := s.publisher.SendMessage(toPBOrder(order), consts.TOPIC_CREATE_ORDER); err != nil {
		s.logger.Error("failed to publish order created event", zap.Error(err))
	}

	// 11) Response
	return &orderpb.CreateOrderResponse{Order: toPBOrder(order)}, nil
}

func (s *Service) GetOrder(ctx context.Context, req *orderpb.GetOrderRequest) (*orderpb.Order, error) {
	accessToken, err := bearerFromMD(ctx)
	if err != nil {
		return nil, err
	}
	valid, userID, role, err := s.authClient.Validate(ctx, accessToken)
	if err != nil || !valid {
		return nil, fmt.Errorf("unauthorised")
	}

	ord, err := s.repo.Get(ctx, req.OrderId) // VALUE: domain.Order
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("order not found")
		}
		return nil, err
	}
	if role == "STAFF" && ord.StaffID != userID {
		return nil, fmt.Errorf("forbidden")
	}
	return toPBOrder(ord), nil // <-- phải &ord
}

func (s *Service) ListOrders(ctx context.Context, req *orderpb.ListOrdersRequest) (*orderpb.ListOrdersResponse, error) {
	accessToken, err := bearerFromMD(ctx)
	if err != nil {
		return nil, err
	}
	valid, userID, role, err := s.authClient.Validate(ctx, accessToken)
	if err != nil || !valid {
		return nil, fmt.Errorf("unauthorised")
	}

	page := req.GetPage()
	limit := req.GetLimit()
	if limit == 0 {
		limit = 10
	}

	orders, total, err := s.repo.List(ctx, page, limit) // VALUE: []domain.Order
	if err != nil {
		return nil, fmt.Errorf("failed to list orders: %w", err)
	}

	filtered := make([]domain.Order, 0, len(orders))
	if role == "STAFF" {
		for _, o := range orders {
			if o.StaffID == userID {
				filtered = append(filtered, o)
			}
		}
	} else {
		filtered = orders
	}

	pbOrders := make([]*orderpb.Order, 0, len(filtered))
	for i := range filtered {
		pbOrders = append(pbOrders, toPBOrder(&filtered[i])) // đúng cách lấy địa chỉ phần tử
	}

	hasNext := int32((page+1)*limit) < total
	pagi := &orderpb.PaginationResponse{
		Total:   total,
		Limit:   limit,
		Page:    page,
		HasNext: hasNext,
	}
	return &orderpb.ListOrdersResponse{Orders: pbOrders, Pagination: pagi}, nil
}

func (s *Service) GenerateInvoice(ctx context.Context, req *orderpb.GetOrderRequest) (*orderpb.GenerateInvoiceResponse, error) {
	// Auth
	accessToken, err := bearerFromMD(ctx)
	if err != nil {
		return nil, err
	}
	valid, userID, role, err := s.authClient.Validate(ctx, accessToken)
	if err != nil || !valid {
		return nil, fmt.Errorf("unauthorised")
	}

	// Fetch order (VALUE)
	ord, err := s.repo.Get(ctx, req.GetOrderId())
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("order not found")
		}
		return nil, err
	}
	// RBAC
	if role == "STAFF" && ord.StaffID != userID {
		return nil, fmt.Errorf("forbidden")
	}

	// PDF (ASCII-only text — no Unicode)
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()

	// Header
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, "INVOICE")
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 11)
	pdf.Cell(0, 7, fmt.Sprintf("Order #: %d", ord.OrderID))
	pdf.Ln(6)
	pdf.Cell(0, 7, fmt.Sprintf("Date: %s", ord.CreatedAt.Format("2006-01-02 15:04")))
	pdf.Ln(6)
	pdf.Cell(0, 7, fmt.Sprintf("Staff: %s", ord.StaffID))
	pdf.Ln(6)
	if ord.CustomerName != "" {
		pdf.Cell(0, 7, fmt.Sprintf("Customer: %s (ID: %d)", ord.CustomerName, ord.CustomerID))
	} else {
		pdf.Cell(0, 7, fmt.Sprintf("Customer ID: %d", ord.CustomerID))
	}
	pdf.Ln(10)

	// Table header
	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(15, 8, "ID", "1", 0, "C", false, 0, "")
	pdf.CellFormat(85, 8, "Item", "1", 0, "L", false, 0, "")
	pdf.CellFormat(20, 8, "Qty", "1", 0, "C", false, 0, "")
	pdf.CellFormat(35, 8, "Unit Price", "1", 0, "R", false, 0, "")
	pdf.CellFormat(35, 8, "Line Total", "1", 0, "R", false, 0, "")
	pdf.Ln(-1)

	// Rows
	pdf.SetFont("Arial", "", 10)
	for _, it := range ord.Items {
		name := it.ProductName
		if name == "" {
			name = fmt.Sprintf("Product #%d", it.ProductID)
		}
		pdf.CellFormat(15, 8, fmt.Sprintf("%d", it.ProductID), "1", 0, "C", false, 0, "")
		pdf.CellFormat(85, 8, name, "1", 0, "L", false, 0, "")
		pdf.CellFormat(20, 8, fmt.Sprintf("%d", it.Quantity), "1", 0, "C", false, 0, "")
		pdf.CellFormat(35, 8, formatVNDEn(it.UnitPrice), "1", 0, "R", false, 0, "")
		pdf.CellFormat(35, 8, formatVNDEn(it.LineTotal), "1", 0, "R", false, 0, "")
		pdf.Ln(-1)
	}

	// Totals
	pdf.Ln(2)
	pdf.SetFont("Arial", "", 11)
	right := func(label, value string) {
		pdf.CellFormat(155, 7, label, "", 0, "R", false, 0, "")
		pdf.CellFormat(35, 7, value, "1", 1, "R", false, 0, "")
	}
	right("Subtotal:", formatVNDEn(ord.TotalPrice))
	right("Shipping:", formatVNDEn(ord.ShippingCost))
	right("Discount:", "- "+formatVNDEn(ord.DiscountAmount))
	pdf.SetFont("Arial", "B", 12)
	right("TOTAL:", formatVNDEn(ord.FinalPrice))

	// Footer
	pdf.Ln(8)
	pdf.SetFont("Arial", "I", 10)
	pdf.Cell(0, 6, "Thank you for your purchase.")

	// Output
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("failed to render pdf: %w", err)
	}

	return &orderpb.GenerateInvoiceResponse{
		FileName: fmt.Sprintf("invoice_%d.pdf", ord.OrderID),
		FileData: buf.Bytes(),
	}, nil
}

// ===== helpers =====

func bearerFromMD(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", fmt.Errorf("missing authorization token")
	}
	vals := md["authorization"]
	if len(vals) == 0 {
		return "", fmt.Errorf("missing authorization token")
	}
	authHeader := vals[0]
	if !strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		return "", fmt.Errorf("missing authorization token")
	}
	return strings.TrimSpace(authHeader[7:]), nil
}

func toPBOrder(o *domain.Order) *orderpb.Order {
	pb := &orderpb.Order{
		OrderId:        o.OrderID,
		CustomerName:   o.CustomerName,
		CustomerId:     o.CustomerID, // NEW
		StaffId:        o.StaffID,
		VoucherCodes:   o.VoucherCodes,
		TotalPrice:     o.TotalPrice,
		DiscountAmount: o.DiscountAmount,
		FinalPrice:     o.FinalPrice,
		ShippingCost:   o.ShippingCost,
		CreatedAt:      timestamppb.New(o.CreatedAt),
		Status:         orderStatusDomainToPB(o.Status),
	}
	for _, it := range o.Items {
		pb.Items = append(pb.Items, &orderpb.OrderItem{
			ProductId:    it.ProductID,
			Quantity:     it.Quantity,
			UnitPrice:    it.UnitPrice,
			ProductName:  it.ProductName,
			ProductImage: it.ProductImage,
			LineTotal:    it.LineTotal,
		})
	}
	for _, h := range o.StatusHistory {
		pb.StatusHistory = append(pb.StatusHistory, &orderpb.StatusHistory{
			Status: orderStatusDomainToPB(h.Status),
			Note:   h.Note,
			At:     timestamppb.New(h.At),
		})
	}
	return pb
}

func orderStatusDomainToPB(s domain.OrderStatus) orderpb.OrderStatus {
	switch s {
	case domain.OrderStatusPending:
		return orderpb.OrderStatus_PENDING
	case domain.OrderStatusPaid:
		return orderpb.OrderStatus_PAID
	case domain.OrderStatusCompleted:
		return orderpb.OrderStatus_COMPLETED
	case domain.OrderStatusCanceled:
		return orderpb.OrderStatus_CANCELED
	default:
		return orderpb.OrderStatus_ORDER_STATUS_UNSPECIFIED
	}
}

func formatVNDEn(n float64) string {
	// Round to integer VND
	i := int64(n + 0.5)
	s := strconv.FormatInt(i, 10)
	if len(s) == 0 {
		return "0 VND"
	}

	// Insert thousand-separator commas from right to left
	var out []byte
	for j, c := range []byte(s) {
		// Calculate position from the right
		posFromRight := len(s) - j - 1
		if j > 0 && posFromRight%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, c)
	}
	return string(out) + " VND"
}
