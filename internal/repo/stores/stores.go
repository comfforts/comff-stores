package stores

import (
	"context"
	"errors"
	"regexp"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otelcodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/comfforts/logger"

	indom "github.com/comfforts/comff-stores/internal/domain/infra"
	stdom "github.com/comfforts/comff-stores/internal/domain/stores"
	"github.com/comfforts/comff-stores/internal/infra/observability"
)

const STORES_COLLECTION = "stores.stores"

const (
	ERR_MISSING_REQUIRED = "missing required parameters"
	ERR_DUPLICATE_STORE  = "duplicate store"
	ERR_DECODING_REC_ID  = "error decoding record ID"
	ERR_NO_STORE         = "no store found"
)

var (
	ErrMissingRequired = errors.New(ERR_MISSING_REQUIRED)
	ErrDuplicateStore  = errors.New(ERR_DUPLICATE_STORE)
	ErrDecodeRecId     = errors.New(ERR_DECODING_REC_ID)
	ErrNoStore         = errors.New(ERR_NO_STORE)
)

type storesRepo struct {
	indom.DBStore
	metrics observability.Metrics
}

func NewStoresRepo(ctx context.Context, rc indom.DBStore, mt observability.Metrics) (*storesRepo, error) {
	l, err := logger.LoggerFromContext(ctx)
	if err != nil {
		l = logger.GetSlogLogger()
	}

	// ensure stores indexes
	if err = rc.EnsureIndexes(ctx, STORES_COLLECTION, []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "address_id", Value: 1},
			},
			Options: options.Index().SetUnique(true), // unique index on address_id
		},
	}); err != nil {
		l.Error("error adding stores indexes", "error", err.Error())
		return nil, err
	}

	l.Info("initialized stores repo")
	return &storesRepo{
		DBStore: rc,
		metrics: mt,
	}, nil
}

func (sr *storesRepo) AddStore(ctx context.Context, st *stdom.Store) (string, error) {
	ctx, span := startSpan(ctx, "stores.repo.add")
	defer span.End()

	l, err := logger.LoggerFromContext(ctx)
	if err != nil {
		l = logger.GetSlogLogger()
	}
	l.Debug("adding store")

	if st == nil || st.AddressId == "" || st.Name == "" || st.Org == "" {
		finishSpan(span, ErrMissingRequired)
		return "", ErrMissingRequired
	}

	coll := sr.Store().Collection(STORES_COLLECTION)

	res, err := coll.InsertOne(ctx, st)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			finishSpan(span, ErrDuplicateStore)
			return "", ErrDuplicateStore
		}

		finishSpan(span, err)
		return "", err
	}

	id, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		finishSpan(span, ErrDecodeRecId)
		return "", ErrDecodeRecId
	}
	return id.Hex(), nil
}

func (sr *storesRepo) GetStore(ctx context.Context, idHex string) (*stdom.Store, error) {
	ctx, span := startSpan(ctx, "stores.repo.get")
	defer span.End()

	l, err := logger.LoggerFromContext(ctx)
	if err != nil {
		l = logger.GetSlogLogger()
	}
	l.Debug("getting store")

	if idHex == "" {
		finishSpan(span, ErrMissingRequired)
		return nil, ErrMissingRequired
	}

	coll := sr.Store().Collection(STORES_COLLECTION)
	objID, err := primitive.ObjectIDFromHex(idHex)
	if err != nil {
		l.Error("GetStore error invalid idHex", "error", err.Error())
		finishSpan(span, ErrDecodeRecId)
		return nil, ErrDecodeRecId
	}
	filter := bson.M{"_id": objID}

	var store stdom.Store
	err = coll.FindOne(ctx, filter).Decode(&store)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			finishSpan(span, ErrNoStore)
			return nil, ErrNoStore
		}
		l.Error("GetStore error", "error", err.Error())
		finishSpan(span, err)
		return nil, err
	}
	return &store, nil
}

func (sr *storesRepo) DeleteStore(ctx context.Context, idHex string) error {
	ctx, span := startSpan(ctx, "stores.repo.delete")
	defer span.End()

	l, err := logger.LoggerFromContext(ctx)
	if err != nil {
		l = logger.GetSlogLogger()
	}
	l.Debug("deleting store")

	if idHex == "" {
		finishSpan(span, ErrMissingRequired)
		return ErrMissingRequired
	}

	coll := sr.Store().Collection(STORES_COLLECTION)
	objID, err := primitive.ObjectIDFromHex(idHex)
	if err != nil {
		l.Error("DeleteStore error invalid idHex", "error", err.Error())
		finishSpan(span, ErrDecodeRecId)
		return ErrDecodeRecId
	}
	filter := bson.M{"_id": objID}

	res, err := coll.DeleteOne(ctx, filter)
	if err != nil {
		l.Error("DeleteStore error", "error", err.Error())
		finishSpan(span, err)
		return err
	}
	if res.DeletedCount == 0 {
		finishSpan(span, ErrNoStore)
		return ErrNoStore
	}

	return nil
}

func (sr *storesRepo) UpdateStore(ctx context.Context, idHex string, params *stdom.UpdateStoreQuery) error {
	ctx, span := startSpan(ctx, "stores.repo.update")
	defer span.End()

	l, err := logger.LoggerFromContext(ctx)
	if err != nil {
		l = logger.GetSlogLogger()
	}
	l.Debug("updating store")

	if idHex == "" {
		finishSpan(span, ErrMissingRequired)
		return ErrMissingRequired
	}

	coll := sr.Store().Collection(STORES_COLLECTION)
	objID, err := primitive.ObjectIDFromHex(idHex)
	if err != nil {
		l.Error("UpdateStore error invalid idHex", "error", err.Error())
		finishSpan(span, ErrDecodeRecId)
		return ErrDecodeRecId
	}
	filter := bson.M{"_id": objID}

	updateParams := bson.M{}
	if params.Name != "" {
		updateParams["name"] = params.Name
	}
	if params.Org != "" {
		updateParams["org"] = params.Org
	}
	if params.AddressId != "" {
		updateParams["address_id"] = params.AddressId
	}
	if len(updateParams) == 0 {
		finishSpan(span, ErrMissingRequired)
		return ErrMissingRequired
	}
	update := bson.M{"$set": updateParams}

	res, err := coll.UpdateOne(ctx, filter, update)
	if err != nil {
		l.Error("UpdateStore error", "error", err.Error())
		finishSpan(span, err)
		return err
	}
	if res.MatchedCount == 0 {
		finishSpan(span, ErrNoStore)
		return ErrNoStore
	}

	return nil
}

func (sr *storesRepo) SearchStores(ctx context.Context, params *stdom.SearchStoreQuery) ([]*stdom.Store, error) {
	ctx, span := startSpan(ctx, "stores.repo.search")
	defer span.End()

	l, err := logger.LoggerFromContext(ctx)
	if err != nil {
		l = logger.GetSlogLogger()
	}
	l.Debug("searching stores")

	coll := sr.Store().Collection(STORES_COLLECTION)
	filter := bson.M{}
	if params.Org != "" {
		filter["org"] = bson.M{"$regex": "^" + regexp.QuoteMeta(params.Org), "$options": "i"}
	}
	if params.Name != "" {
		filter["name"] = bson.M{"$regex": "^" + regexp.QuoteMeta(params.Name), "$options": "i"}
	}
	if params.AddressId != "" {
		filter["address_id"] = bson.M{"$regex": "^" + regexp.QuoteMeta(params.AddressId), "$options": "i"}
	}

	cursor, err := coll.Find(ctx, filter)
	if err != nil {
		l.Error("SearchStores error", "error", err.Error())
		finishSpan(span, err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var storesList []*stdom.Store
	for cursor.Next(ctx) {
		var st stdom.Store
		if err := cursor.Decode(&st); err != nil {
			l.Error("SearchStores error decoding store", "error", err.Error())
			continue
		}
		storesList = append(storesList, &st)
	}

	if err := cursor.Err(); err != nil {
		l.Error("SearchStores cursor error", "error", err.Error())
		finishSpan(span, err)
		return nil, err
	}

	return storesList, nil
}

func (sr *storesRepo) Close(ctx context.Context) error {
	return sr.DBStore.Close(ctx)
}

func startSpan(ctx context.Context, name string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	return otel.Tracer("stores-repo").Start(ctx, name, trace.WithAttributes(attrs...))
}

func finishSpan(span trace.Span, err error) {
	if err == nil {
		return
	}
	span.RecordError(err)
	span.SetStatus(otelcodes.Error, err.Error())
}
