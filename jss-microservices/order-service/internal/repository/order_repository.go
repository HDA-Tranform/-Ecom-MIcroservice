package repository

import (
    "context"
    "errors"
    "time"

    "github.com/linhhuynhcoding/jss-microservices/order-service/internal/domain"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

// ErrNotFound is returned when an order cannot be located in the database.
var ErrNotFound = errors.New("order not found")

// OrderRepository provides CRUD operations for orders stored in MongoDB.
// It also manages a counter collection used to generate sequential order
// identifiers.  All methods accept a context to allow cancellation
// propagation.
type OrderRepository struct {
    coll       *mongo.Collection
    counters   *mongo.Collection
}

// New creates a new OrderRepository.  The provided db should be a
// connected mongo.Database; the repository will use two collections: one
// called "orders" and another called "counters" for sequence generation.
func New(db *mongo.Database) *OrderRepository {
    return &OrderRepository{
        coll:     db.Collection("orders"),
        counters: db.Collection("counters"),
    }
}

// NextOrderID atomically increments and returns the next order ID.  It
// relies on a document in the counters collection with _id="orderId".
func (r *OrderRepository) NextOrderID(ctx context.Context) (int32, error) {
    var res struct {
        Seq int32 `bson:"seq"`
    }
    // Use findOneAndUpdate with upsert to ensure the counter exists.  The
    // return document should contain the updated value.
    err := r.counters.FindOneAndUpdate(
        ctx,
        bson.M{"_id": "orderId"},
        bson.M{"$inc": bson.M{"seq": 1}},
        options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After),
    ).Decode(&res)
    if err != nil {
        return 0, err
    }
    return res.Seq, nil
}

// Create inserts a new order document into the orders collection.  The
// CreatedAt field will be set to the current time if not already
// populated.  This method returns any error returned by the driver.
func (r *OrderRepository) Create(ctx context.Context, order *domain.Order) error {
    if order.CreatedAt.IsZero() {
        order.CreatedAt = time.Now()
    }
    _, err := r.coll.InsertOne(ctx, order)
    return err
}

// Get retrieves an order by its sequential order_id.  If the order is
// not found, ErrNotFound is returned.
func (r *OrderRepository) Get(ctx context.Context, orderID int32) (*domain.Order, error) {
    var order domain.Order
    err := r.coll.FindOne(ctx, bson.M{"order_id": orderID}).Decode(&order)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return nil, ErrNotFound
        }
        return nil, err
    }
    return &order, nil
}

// List returns a slice of orders using pagination.  Orders are sorted by
// created_at descending so that the most recent orders appear first.  No
// filtering is applied at this layer; callers may post‑filter the results
// based on role or staff ID.
func (r *OrderRepository) List(ctx context.Context, page, limit int32) ([]domain.Order, int32, error) {
    opts := options.Find().SetSkip(int64(page * limit)).SetLimit(int64(limit)).SetSort(bson.D{{Key: "created_at", Value: -1}})
    cursor, err := r.coll.Find(ctx, bson.M{}, opts)
    if err != nil {
        return nil, 0, err
    }
    var orders []domain.Order
    if err := cursor.All(ctx, &orders); err != nil {
        return nil, 0, err
    }
    // Count total documents for pagination metadata
    count, err := r.coll.CountDocuments(ctx, bson.M{})
    if err != nil {
        return nil, 0, err
    }
    return orders, int32(count), nil
}