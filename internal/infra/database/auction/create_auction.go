package auction

import (
	"context"
	"fullcycle-auction_go/configuration/logger"
	"fullcycle-auction_go/internal/entity/auction_entity"
	"fullcycle-auction_go/internal/internal_error"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type AuctionEntityMongo struct {
	Id          string                          `bson:"_id"`
	ProductName string                          `bson:"product_name"`
	Category    string                          `bson:"category"`
	Description string                          `bson:"description"`
	Condition   auction_entity.ProductCondition `bson:"condition"`
	Status      auction_entity.AuctionStatus    `bson:"status"`
	Timestamp   int64                           `bson:"timestamp"`
}
type AuctionRepository struct {
	Collection *mongo.Collection
}

func NewAuctionRepository(database *mongo.Database) *AuctionRepository {
	return &AuctionRepository{
		Collection: database.Collection("auctions"),
	}
}

func (ar *AuctionRepository) CreateAuction(
	ctx context.Context,
	auctionEntity *auction_entity.Auction) *internal_error.InternalError {
	auctionEntityMongo := &AuctionEntityMongo{
		Id:          auctionEntity.Id,
		ProductName: auctionEntity.ProductName,
		Category:    auctionEntity.Category,
		Description: auctionEntity.Description,
		Condition:   auctionEntity.Condition,
		Status:      auctionEntity.Status,
		Timestamp:   auctionEntity.Timestamp.Unix(),
	}
	opts, err := ar.Collection.InsertOne(ctx, auctionEntityMongo)
	if err != nil {
		logger.Error("Error trying to insert auction", err)
		return internal_error.NewInternalServerError("Error trying to insert auction")
	}

	ar.setAuctionTimeout(ctx, opts.InsertedID.(string))

	return nil
}

func (ar *AuctionRepository) setAuctionTimeout(ctx context.Context, auctionId string) {

	go func(ctx context.Context, ar *AuctionRepository, auctionId string) {
		auctionTimeout := os.Getenv("AUCTION_TIMEOUT")
		duration, err := time.ParseDuration(auctionTimeout)
		if err != nil {
			log.Fatal("Ocorreu um erro ao converter o timeout do arquivo .env", err)
		}
		time.Sleep(duration)

		ar.Collection.UpdateOne(ctx,
			bson.M{"_id": auctionId},
			bson.M{"$set": bson.M{"status": 1}},
		)

	}(ctx, ar, auctionId)

}
