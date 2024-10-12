package mongodb

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/event"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestCRUD(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	monitor := &event.CommandMonitor{
		Started: func(ctx context.Context, startedEvent *event.CommandStartedEvent) {
			fmt.Println(startedEvent.Command)
		},
	}
	opts := options.Client().ApplyURI("mongodb://root:example@localhost:27017/").SetMonitor(monitor)
	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		panic(err)
	}
	defer func() {
		// 用完就是要记得关掉。正常来说，都是在应用退出的时候关掉。
		_ = client.Disconnect(context.Background())
	}()

	col := client.Database("webook").Collection("articles")
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	res, err := col.InsertOne(ctx, Article{
		Id:       123,
		Title:    "我的标题",
		Content:  "我的内容",
		AuthorId: 12,
		Status:   1,
		Ctime:    time.Now().UnixMilli(),
		Utime:    time.Now().UnixMilli(),
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("插入ID %s \n", res.InsertedID)

	// 用 bson 来构造查询条件
	filter := bson.D{bson.E{Key: "id", Value: 123}}
	findRes := col.FindOne(ctx, filter)

	var art Article
	err = findRes.Decode(&art)
	if err != nil {
		panic(err)
	}

	findRes = col.FindOne(ctx, Article{Id: 123})
	if findRes.Err() == mongo.ErrNoDocuments {
		// 这边查询不到
		fmt.Println(err)
	}

	// 只更新标题字段
	sets := bson.D{bson.E{Key: "$set",
		Value: bson.E{Key: "title", Value: "新的标题"}}}
	updateOneRes, err := col.UpdateOne(ctx, filter, sets)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Update One 更新了 %d 条数据 \n", updateOneRes.MatchedCount)

	updateManyRes, err := col.UpdateMany(ctx, filter,
		bson.D{bson.E{Key: "$set", Value: Article{
			Id:    123,
			Title: "直接使用文档更新",
		}}})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Update Many 更新了 %d 条数据 \n", updateManyRes.MatchedCount)

	findRes = col.FindOne(ctx, filter)
	art = Article{}
	err = findRes.Decode(&art)
	if err != nil {
		panic(err)
	}
	delRes, err := col.DeleteMany(ctx, filter)
	if err != nil {
		panic(err)
	}
	fmt.Printf("删除了 %d 条数据\n", delRes.DeletedCount)
}

func TestOr(t *testing.T) {
}

type Article struct {
	Id       int64  `bson:"id"`
	Title    string `bson:"title,omitempty"`
	Content  string `bson:"content,omitempty"` // omitempty 表示如果该数据为零值，则忽略
	AuthorId int64  `bson:"authorId,omitempty"`
	Status   uint8  `bson:"status,omitempty"`
	Ctime    int64  `bson:"ctime,omitempty"`
	Utime    int64  `bson:"utime,omitempty"`
}
