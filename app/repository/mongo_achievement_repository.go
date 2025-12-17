package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoAchievement struct {
	ID              primitive.ObjectID `bson:"_id"`
	AchievementType string             `bson:"achievementType"`
	Details         bson.M             `bson:"details"`
	Tags            []string           `bson:"tags"`
}

type MongoAchievementRepository struct {
	Collection *mongo.Collection
}

type Achievement struct {
	AchievementType string `bson:"achievementType"`

	Details struct {
		Year             int    `bson:"year"`
		CompetitionLevel string `bson:"competitionLevel"`
	} `bson:"details"`
}

func NewMongoAchievementRepository(col *mongo.Collection) *MongoAchievementRepository {
	return &MongoAchievementRepository{
		Collection: col,
	}
}

// FindByIDs mengambil banyak achievement dari MongoDB berdasarkan daftar ID
func (r *MongoAchievementRepository) FindByIDs(
	ctx context.Context,
	ids []primitive.ObjectID,
) ([]Achievement, error) {

	filter := bson.M{"_id": bson.M{"$in": ids}}

	cursor, err := r.Collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []Achievement
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

