package service

import (
	"database/sql"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	utils "github.com/linhhuynhcoding/jss-microservices/jss-shared/utils/format"
	db "github.com/linhhuynhcoding/jss-microservices/loyalty/internal/repository"
	api "github.com/linhhuynhcoding/jss-microservices/rpc/gen/loyalty"
)

// ===== NULLABLE TYPE CONVERTERS =====

// ConvertOptionalInt32ToNullInt32 converts an optional int32 pointer to sql.NullInt32
func ConvertOptionalInt32ToNullInt32(value *int32) sql.NullInt32 {
	if value == nil {
		return sql.NullInt32{Valid: false}
	}
	return sql.NullInt32{Int32: *value, Valid: true}
}

// ConvertNullInt32ToOptionalInt32 converts sql.NullInt32 to optional int32 pointer
func ConvertNullInt32ToOptionalInt32(value sql.NullInt32) *int32 {
	if !value.Valid {
		return nil
	}
	return &value.Int32
}

// ConvertOptionalStringToNullString converts an optional string pointer to sql.NullString
func ConvertOptionalStringToNullString(value *string) sql.NullString {
	if value == nil {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: *value, Valid: true}
}

// ConvertNullStringToOptionalString converts sql.NullString to optional string pointer
func ConvertNullStringToOptionalString(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}

// ConvertTimeToTimestamp converts time.Time to timestamppb.Timestamp
func ConvertTimeToTimestamp(t time.Time) *timestamppb.Timestamp {
	return timestamppb.New(t)
}

// ConvertOptionalTimeToTimestamp converts optional time.Time to timestamppb.Timestamp
func ConvertOptionalTimeToTimestamp(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}
	return timestamppb.New(*t)
}

// ===== DISCOUNT TYPE CONVERTERS =====

// ConvertAPIDiscountTypeToDBString converts API DiscountType enum to database string
func ConvertAPIDiscountTypeToDBString(discountType api.Voucher_DiscountType) string {
	switch discountType {
	case api.Voucher_PERCENTAGE:
		return "PERCENTAGE"
	case api.Voucher_FIXED_AMOUNT:
		return "FIXED_AMOUNT"
	case api.Voucher_FREE_SHIPPING:
		return "FREE_SHIPPING"
	default:
		return "PERCENTAGE" // default fallback
	}
}

// ConvertDBStringToAPIDiscountType converts database string to API DiscountType enum
func ConvertDBStringToAPIDiscountType(discountType string) api.Voucher_DiscountType {
	switch discountType {
	case "PERCENTAGE":
		return api.Voucher_PERCENTAGE
	case "FIXED_AMOUNT":
		return api.Voucher_FIXED_AMOUNT
	case "FREE_SHIPPING":
		return api.Voucher_FREE_SHIPPING
	default:
		return api.Voucher_DISCOUNT_TYPE_UNSPECIFIED
	}
}

// ===== CUSTOMER VOUCHER STATUS CONVERTERS =====

// ConvertAPICustomerVoucherStatusToDBString converts API CustomerVoucher Status enum to database string
func ConvertAPICustomerVoucherStatusToDBString(status *api.CustomerVoucher_Status) string {
	if status == nil {
		return "UNUSED" // default value
	}

	switch *status {
	case api.CustomerVoucher_UNUSED:
		return "UNUSED"
	case api.CustomerVoucher_USED:
		return "USED"
	case api.CustomerVoucher_EXPIRED:
		return "EXPIRED"
	case api.CustomerVoucher_CANCELLED:
		return "CANCELLED"
	default:
		return "UNUSED" // default fallback
	}
}

// ConvertDBStringToAPICustomerVoucherStatus converts database string to API CustomerVoucher Status enum
func ConvertDBStringToAPICustomerVoucherStatus(status string) api.CustomerVoucher_Status {
	switch status {
	case "UNUSED":
		return api.CustomerVoucher_UNUSED
	case "USED":
		return api.CustomerVoucher_USED
	case "EXPIRED":
		return api.CustomerVoucher_EXPIRED
	case "CANCELLED":
		return api.CustomerVoucher_CANCELLED
	default:
		return api.CustomerVoucher_STATUS_UNSPECIFIED
	}
}

// ===== ENTITY CONVERTERS =====

// ConvertDBLoyaltyPointToAPI converts database LoyaltyPoint to API LoyaltyPoint
func ConvertDBLoyaltyPointToAPI(dbPoint db.LoyaltyPoint) *api.LoyaltyPoint {
	return &api.LoyaltyPoint{
		Id:         dbPoint.ID,
		CustomerId: dbPoint.CustomerID,
		Points:     dbPoint.Points.Int32,
		Source:     dbPoint.Source,
		// ReferenceId: ConvertNullInt32ToOptionalInt32(dbPoint.ReferenceID),
		CreatedAt: utils.PgToPbTimestamp(dbPoint.CreatedAt),
	}
}

// ConvertDBVoucherToAPI converts database Voucher to API Voucher
func ConvertDBVoucherToAPI(dbVoucher db.Voucher, used *int32) *api.Voucher {
	return &api.Voucher{
		Id:            dbVoucher.ID,
		Code:          dbVoucher.Code,
		Description:   utils.PgTextToStringPointer(dbVoucher.Description),
		DiscountType:  ConvertDBStringToAPIDiscountType(dbVoucher.DiscountType),
		DiscountValue: utils.NumericToFloat64(dbVoucher.DiscountValue),
		StartDate:     utils.PgDateToString(dbVoucher.StartDate),
		EndDate:       utils.PgDateToString(dbVoucher.EndDate),
		UsageLimit:    &dbVoucher.UsageLimit.Int32,
		CreatedAt:     utils.PgToPbTimestamp(dbVoucher.CreatedAt),
		UsedNumber:    used, // Assuming this field exists in db.Voucher
	}
}

// ConvertDBCustomerVoucherToAPI converts database CustomerVoucher to API CustomerVoucher
func ConvertDBCustomerVoucherToAPI(dbCustomerVoucher db.CustomerVoucher) *api.CustomerVoucher {
	return &api.CustomerVoucher{
		Id:         dbCustomerVoucher.ID,
		CustomerId: dbCustomerVoucher.CustomerID,
		VoucherId:  dbCustomerVoucher.VoucherID,
		Status:     ConvertDBStringToAPICustomerVoucherStatus(dbCustomerVoucher.Status.String),
		UsedAt:     utils.PgToPbTimestamp(dbCustomerVoucher.UsedAt),
		Voucher:    nil, // Will be populated separately if needed
	}
}

// ConvertDBCustomerVoucherRowToAPI converts database GetCustomerVouchersRow to API CustomerVoucher
func ConvertDBCustomerVoucherRowToAPI(dbRow db.GetCustomerVouchersRow) *api.CustomerVoucher {
	customerVoucher := &api.CustomerVoucher{
		Id:         dbRow.ID,
		CustomerId: dbRow.CustomerID,
		VoucherId:  dbRow.VoucherID,
		Status:     ConvertDBStringToAPICustomerVoucherStatus(dbRow.Status.String),
		UsedAt:     utils.PgToPbTimestamp(dbRow.UsedAt),
	}

	// If voucher details are included in the row, populate them
	customerVoucher.Voucher = &api.Voucher{
		Id:            dbRow.VoucherID,
		Code:          dbRow.Code,
		Description:   ConvertNullStringToOptionalString(sql.NullString(dbRow.Description)),
		DiscountType:  ConvertDBStringToAPIDiscountType(dbRow.DiscountType),
		DiscountValue: utils.NumericToFloat64(dbRow.DiscountValue),
		StartDate:     utils.PgDateToString(dbRow.StartDate),
		EndDate:       utils.PgDateToString(dbRow.EndDate),
		// UsageLimit:    ConvertNullInt32ToOptionalInt32(dbRow.),
		// CreatedAt:     utils.PgToPbTimestamp(dbRow),
	}

	return customerVoucher
}

// ConvertDBCustomerVoucherStatusRowToAPI converts database GetCustomerVouchersByStatusRow to API CustomerVoucher
func ConvertDBCustomerVoucherStatusRowToAPI(dbRow db.GetCustomerVouchersByStatusRow) *api.CustomerVoucher {
	customerVoucher := &api.CustomerVoucher{
		Id:         dbRow.ID,
		CustomerId: dbRow.CustomerID,
		VoucherId:  dbRow.VoucherID,
		Status:     ConvertDBStringToAPICustomerVoucherStatus(dbRow.Status.String),
		UsedAt:     utils.PgToPbTimestamp(dbRow.UsedAt),
	}

	// If voucher details are included in the row, populate them
	customerVoucher.Voucher = &api.Voucher{
		Id:            dbRow.VoucherID,
		Code:          dbRow.Code,
		Description:   ConvertNullStringToOptionalString(sql.NullString(dbRow.Description)),
		DiscountType:  ConvertDBStringToAPIDiscountType(dbRow.DiscountType),
		DiscountValue: utils.NumericToFloat64(dbRow.DiscountValue),
		StartDate:     utils.PgDateToString(dbRow.StartDate),
		EndDate:       utils.PgDateToString(dbRow.EndDate),
		// UsageLimit:    ConvertNullInt32ToOptionalInt32(dbRow.VoucherUsageLimit),
		// CreatedAt:     utils.PgToPbTimestamp(dbRow.VoucherCreatedAt),
	}

	return customerVoucher
}

// ConvertDBAllCustomerVouchersRowToAPI converts database GetAllCustomerVouchersRow to API CustomerVoucher
func ConvertDBAllCustomerVouchersRowToAPI(dbRow db.GetAllCustomerVouchersRow) *api.CustomerVoucher {
	customerVoucher := &api.CustomerVoucher{
		Id:         dbRow.ID,
		CustomerId: dbRow.CustomerID,
		VoucherId:  dbRow.VoucherID,
		Status:     ConvertDBStringToAPICustomerVoucherStatus(dbRow.Status.String),
		UsedAt:     utils.PgToPbTimestamp(dbRow.UsedAt),
	}

	// If voucher details are included in the row, populate them
	customerVoucher.Voucher = &api.Voucher{
		Id:            dbRow.VoucherID,
		Code:          dbRow.Code,
		Description:   ConvertNullStringToOptionalString(sql.NullString(dbRow.Description)),
		DiscountType:  ConvertDBStringToAPIDiscountType(dbRow.DiscountType),
		DiscountValue: utils.NumericToFloat64(dbRow.DiscountValue),
		StartDate:     utils.PgDateToString(dbRow.StartDate),
		EndDate:       utils.PgDateToString(dbRow.EndDate),
		// UsageLimit:    ConvertNullInt32ToOptionalInt32(dbRow.VoucherUsageLimit),
		// CreatedAt:     utils.PgToPbTimestamp(dbRow.VoucherCreatedAt),
	}
	return customerVoucher
}

// ===== VALIDATION HELPERS =====

// IsValidDateFormat validates if a string is in YYYY-MM-DD format
func IsValidDateFormat(dateStr string) bool {
	dateRegex := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	if !dateRegex.MatchString(dateStr) {
		return false
	}

	// Try to parse the date to ensure it's valid
	_, err := time.Parse("2006-01-02", dateStr)
	return err == nil
}

// ===== DISCOUNT CALCULATION HELPERS =====

// CalculatePercentageDiscount calculates percentage-based discount
func CalculatePercentageDiscount(amount, percentage float64) float64 {
	if percentage < 0 || percentage > 100 {
		return 0
	}
	return amount * (percentage / 100.0)
}

// CalculateFixedAmountDiscount calculates fixed amount discount
func CalculateFixedAmountDiscount(amount, discountValue float64) float64 {
	if discountValue <= 0 {
		return 0
	}
	if discountValue > amount {
		return amount
	}
	return discountValue
}

// CalculateVoucherDiscount calculates discount amount based on voucher type
func CalculateVoucherDiscount(voucher *api.Voucher, productAmount, shippingAmount float64) float64 {
	switch voucher.DiscountType {
	case api.Voucher_PERCENTAGE:
		return CalculatePercentageDiscount(productAmount, voucher.DiscountValue)
	case api.Voucher_FIXED_AMOUNT:
		return CalculateFixedAmountDiscount(productAmount, voucher.DiscountValue)
	case api.Voucher_FREE_SHIPPING:
		return shippingAmount
	default:
		return 0
	}
}

// IsVoucherActive checks if a voucher is currently active based on dates
func IsVoucherActive(voucher *api.Voucher) bool {
	now := time.Now()
	currentDate := now.Format("2006-01-02")

	return currentDate >= voucher.StartDate && currentDate <= voucher.EndDate
}

// ===== VOUCHER USAGE HELPERS =====

// CanUseVoucher checks if a voucher can be used based on usage limits
func CanUseVoucher(voucher *api.Voucher, currentUsage int32) bool {
	if voucher.UsageLimit == nil {
		return true // No usage limit
	}
	return currentUsage < *voucher.UsageLimit
}

// ===== PAGINATION HELPERS =====

// CalculateOffset calculates the offset for pagination
func CalculateOffset(page, limit int32) int32 {
	if page <= 0 {
		return 0
	}
	return page * limit
}

// HasNextPage determines if there are more pages available
func HasNextPage(returnedCount, requestedLimit int) bool {
	return returnedCount == requestedLimit
}

// ===== BUSINESS LOGIC CONVERTERS =====

// ConvertDBUsageRecordToAPIVoucherUsage converts database usage records for API response
// func ConvertDBUsageRecordToAPIVoucherUsage(dbRecord db.UsageRecord) *api.CalculateDiscountAmountResponse_Voucher {
// 	return &api.CalculateDiscountAmountResponse_Voucher{
// 		Code:           dbRecord.Code,           // Assuming Code field exists
// 		Title:          dbRecord.VoucherTitle,   // Assuming VoucherTitle field exists
// 		DiscountAmount: dbRecord.DiscountAmount, // Assuming DiscountAmount field exists
// 	}
// }

// ===== VALIDATION FUNCTIONS =====

// ValidateCustomerID validates customer ID
func ValidateCustomerID(customerID int32) error {
	if customerID <= 0 {
		return fmt.Errorf("customer_id must be positive")
	}
	return nil
}

// ValidateVoucherID validates voucher ID
func ValidateVoucherID(voucherID int32) error {
	if voucherID <= 0 {
		return fmt.Errorf("voucher_id must be positive")
	}
	return nil
}

// ValidatePagination validates pagination parameters
func ValidatePagination(pagination *api.PaginationRequest) error {
	if pagination == nil {
		return fmt.Errorf("pagination is required")
	}
	if pagination.Limit <= 0 {
		return fmt.Errorf("limit must be positive")
	}
	if pagination.Limit > 100 {
		return fmt.Errorf("limit cannot exceed 100")
	}
	if pagination.Page < 0 {
		return fmt.Errorf("page cannot be negative")
	}
	return nil
}

// ValidateCode validates voucher code
func ValidateCode(code string) error {
	if code == "" {
		return fmt.Errorf("voucher code is required")
	}
	if len(code) > 50 {
		return fmt.Errorf("voucher code cannot exceed 50 characters")
	}
	return nil
}

// ValidateSource validates loyalty point source
func ValidateSource(source string) error {
	if source == "" {
		return fmt.Errorf("source is required")
	}
	if len(source) > 50 {
		return fmt.Errorf("source cannot exceed 50 characters")
	}
	return nil
}

// ValidateDiscountValue validates discount value based on type
func ValidateDiscountValue(discountType api.Voucher_DiscountType, value float64) error {
	switch discountType {
	case api.Voucher_PERCENTAGE:
		if value < 0 || value > 100 {
			return fmt.Errorf("percentage discount must be between 0 and 100")
		}
	case api.Voucher_FIXED_AMOUNT:
		if value < 0 {
			return fmt.Errorf("fixed amount discount cannot be negative")
		}
	case api.Voucher_FREE_SHIPPING:
		// Free shipping doesn't need value validation
	default:
		return fmt.Errorf("invalid discount type")
	}
	return nil
}

// ===== DATE HELPERS =====

// ParseDateString parses date string in YYYY-MM-DD format
func ParseDateString(dateStr string) (time.Time, error) {
	return time.Parse("2006-01-02", dateStr)
}

// IsDateInRange checks if current date is within the given range
func IsDateInRange(startDate, endDate string) bool {
	now := time.Now()
	currentDate := now.Format("2006-01-02")

	return currentDate >= startDate && currentDate <= endDate
}

// ===== VOUCHER CALCULATION HELPERS =====

// CalculateMultipleVouchersDiscount calculates total discount from multiple vouchers
func CalculateMultipleVouchersDiscount(vouchers []*api.Voucher, productAmount, shippingAmount float64) (float64, []*api.CalculateDiscountAmountResponse_Voucher) {
	var totalDiscount float64
	var voucherDiscounts []*api.CalculateDiscountAmountResponse_Voucher

	remainingProductAmount := productAmount
	remainingShippingAmount := shippingAmount

	for _, voucher := range vouchers {
		if !IsVoucherActive(voucher) {
			continue
		}

		var discountAmount float64

		switch voucher.DiscountType {
		case api.Voucher_PERCENTAGE:
			discountAmount = CalculatePercentageDiscount(remainingProductAmount, voucher.DiscountValue)
			remainingProductAmount -= discountAmount
		case api.Voucher_FIXED_AMOUNT:
			discountAmount = CalculateFixedAmountDiscount(remainingProductAmount, voucher.DiscountValue)
			remainingProductAmount -= discountAmount
		case api.Voucher_FREE_SHIPPING:
			discountAmount = remainingShippingAmount
			remainingShippingAmount = 0
		}

		if discountAmount > 0 {
			totalDiscount += discountAmount
			voucherDiscounts = append(voucherDiscounts, &api.CalculateDiscountAmountResponse_Voucher{
				Code: voucher.Code,
				// Title:          voucher.Description, // Using description as title
				DiscountAmount: discountAmount,
			})
		}

		// Ensure amounts don't go negative
		if remainingProductAmount < 0 {
			remainingProductAmount = 0
		}
		if remainingShippingAmount < 0 {
			remainingShippingAmount = 0
		}
	}

	return totalDiscount, voucherDiscounts
}

// ===== STRING CONVERSION HELPERS =====

// Int32ToString converts int32 to string
func Int32ToString(value int32) string {
	return strconv.FormatInt(int64(value), 10)
}

// StringToInt32 converts string to int32
func StringToInt32(value string) (int32, error) {
	parsed, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(parsed), nil
}

// Float64ToString converts float64 to string
func Float64ToString(value float64) string {
	return strconv.FormatFloat(value, 'f', 2, 64)
}

// StringToFloat64 converts string to float64
func StringToFloat64(value string) (float64, error) {
	return strconv.ParseFloat(value, 64)
}

// ===== ERROR HELPERS =====

// WrapDBError wraps database errors with appropriate gRPC status codes
func WrapDBError(err error, operation string) error {
	if err == nil {
		return nil
	}

	// Check for common database errors and convert to appropriate gRPC codes
	errStr := err.Error()

	if contains(errStr, "not found") || contains(errStr, "no rows") {
		return status.Errorf(codes.NotFound, "%s: record not found", operation)
	}

	if contains(errStr, "duplicate") || contains(errStr, "unique constraint") {
		return status.Errorf(codes.AlreadyExists, "%s: record already exists", operation)
	}

	if contains(errStr, "foreign key") {
		return status.Errorf(codes.FailedPrecondition, "%s: invalid reference", operation)
	}

	return status.Errorf(codes.Internal, "%s: database error", operation)
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
			(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
				indexOf(s, substr) >= 0)))
}

// indexOf returns the index of substr in s, or -1 if not found
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// ===== RESPONSE BUILDERS =====

// BuildPaginationResponse builds a pagination response
func BuildPaginationResponse(totalCount, limit, page int32, hasNext bool) *api.PaginationResponse {
	return &api.PaginationResponse{
		Total:   totalCount,
		Limit:   limit,
		Page:    page,
		HasNext: hasNext,
	}
}

// BuildErrorResponse builds a standardized error response
func BuildErrorResponse(code codes.Code, message string, details ...interface{}) error {
	if len(details) > 0 {
		return status.Errorf(code, message+": %v", details[0])
	}
	return status.Error(code, message)
}
