package mongodb

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"testing"
	"time"
)

type MongoDBTestSuite struct {
	suite.Suite
	col *mongo.Collection
}

func (s *MongoDBTestSuite) SetupSuite() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	monitor := &event.CommandMonitor{
		Started: func(ctx context.Context,
			startedEvent *event.CommandStartedEvent) {
			fmt.Println(startedEvent.Command)
		},
	}
	opts := options.Client().
		ApplyURI("mongodb://root:example@localhost:27017/").
		SetMonitor(monitor)
	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		panic(err)
	}
	s.col = client.Database("webook").Collection("webook")
}

func (s *MongoDBTestSuite) TearDownSuite() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	err := s.col.Database().Client().Disconnect(ctx)
	cancel()
	assert.NoError(s.T(), err)
}

func (s *MongoDBTestSuite) TearDownTest() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	// 删除全部文档
	dres, err := s.col.DeleteMany(ctx, bson.D{})
	require.NoError(s.T(), err)
	s.T().Log(dres.DeletedCount)
}

func (s *MongoDBTestSuite) TestOr() {
	// 先插入一些数据
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	col := s.col
	imany, err := col.InsertMany(ctx, []any{
		Article{
			Id: 123,
		},
		Article{
			Id: 234,
		},
	})
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 2, len(imany.InsertedIDs))

	or := bson.A{bson.D{bson.E{Key: "id", Value: 123}},
		bson.D{bson.E{Key: "id", Value: 234}}}
	res, err := col.Find(ctx, bson.D{bson.E{Key: "$or", Value: or}})
	assert.NoError(s.T(), err)
	var arts []Article
	err = res.All(ctx, &arts)
	assert.NoError(s.T(), err)
	s.T().Log(arts)
}

func (s *MongoDBTestSuite) TestAnd() {
	// 先插入一些数据
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	col := s.col
	imany, err := col.InsertMany(ctx, []any{
		Article{
			Id: 123,
		},
		Article{
			Id:    123,
			Title: "有标题",
		},
	})
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 2, len(imany.InsertedIDs))

	and := bson.A{bson.D{bson.E{Key: "id", Value: 123}},
		bson.D{bson.E{Key: "title", Value: "有标题"}}}
	res, err := col.Find(ctx, bson.D{bson.E{Key: "$and", Value: and}})
	assert.NoError(s.T(), err)
	var arts []Article
	err = res.All(ctx, &arts)
	assert.NoError(s.T(), err)
	s.T().Log(arts)
}

func (s *MongoDBTestSuite) TestIn() {
	// 先插入一些数据
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	col := s.col
	imany, err := col.InsertMany(ctx, []any{
		Article{
			Id: 123,
		},
		Article{
			Id:    234,
			Title: "有标题",
		},
	})
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 2, len(imany.InsertedIDs))

	res, err := col.Find(ctx, bson.D{bson.E{Key: "id",
		Value: bson.D{bson.E{Key: "$in", Value: []int{123, 234}}}}})
	assert.NoError(s.T(), err)
	var arts []Article
	err = res.All(ctx, &arts)
	assert.NoError(s.T(), err)
	s.T().Log(arts)
}

func (s *MongoDBTestSuite) TestProject() {
	// 先插入一些数据
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	col := s.col
	imany, err := col.InsertMany(ctx, []any{
		Article{
			Id: 123,
		},
		Article{
			Id:    234,
			Title: "有标题",
		},
	})
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 2, len(imany.InsertedIDs))

	res, err := col.Find(ctx,
		bson.D{bson.E{Key: "id",
			Value: bson.D{bson.E{Key: "$in", Value: []int{123, 234}}}}},
		// 只查询 id
		options.Find().SetProjection(bson.D{{"id", 1}}))
	assert.NoError(s.T(), err)
	var arts []Article
	err = res.All(ctx, &arts)
	assert.NoError(s.T(), err)
	s.T().Log(arts)
}

func (s *MongoDBTestSuite) TestIndex() {
	// 先插入一些数据
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	col := s.col
	ires, err := s.col.Indexes().
		CreateOne(ctx, mongo.IndexModel{
			Keys: bson.D{bson.E{Key: "id", Value: 1}},
			// 设置为唯一索引
			Options: options.Index().SetUnique(true),
		})
	assert.NoError(s.T(), err)
	s.T().Log(ires)
	_, err = col.InsertMany(ctx, []any{
		Article{
			Id: 123,
		},
		Article{
			Id:    123,
			Title: "有标题",
		},
	})
	assert.NotNil(s.T(), err)
}

func TestMongoDB(t *testing.T) {
	suite.Run(t, &MongoDBTestSuite{})
}
