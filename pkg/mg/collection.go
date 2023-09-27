package mg

import (
	"context"

	"github.com/yusufaine/apple-inventory-notifier/pkg/alert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const AlertsCollection = "alerts"

type Collection struct {
	Context context.Context
	col     *mongo.Collection
}

func NewAlertsConnection(c *Config) *Collection {
	client, err := mongo.Connect(c.Context, options.Client().ApplyURI(c.MongoUri))
	if err != nil {
		panic(err)
	}
	return &Collection{
		Context: c.Context,
		col:     client.Database(c.MongoDb).Collection(AlertsCollection),
	}
}

func (c *Collection) InsertAlerts(alerts *alert.Alerts) int {
	var docs []interface{}
	for _, a := range *alerts {
		docs = append(docs, a.ToBSON())
	}

	res, err := c.col.InsertMany(c.Context, docs)
	if err != nil {
		panic(err)
	}

	return len(res.InsertedIDs)
}

func (c *Collection) GetAlerts() *alert.Alerts {
	cur, err := c.col.Find(c.Context, bson.M{})
	if err != nil {
		panic(err)
	}
	defer cur.Close(c.Context)

	var alerts alert.Alerts
	for cur.Next(c.Context) {
		var alert alert.Alert
		if err := cur.Decode(&alert); err != nil {
			panic(err)
		}
		alerts = append(alerts, alert)
	}

	return &alerts
}

func (c *Collection) GetAlertsByFilter(filter bson.M) *alert.Alerts {
	cur, err := c.col.Find(c.Context, filter)
	if err != nil {
		panic(err)
	}
	defer cur.Close(c.Context)

	var alerts alert.Alerts
	for cur.Next(c.Context) {
		var alert alert.Alert
		if err := cur.Decode(&alert); err != nil {
			panic(err)
		}
		alerts = append(alerts, alert)
	}

	return &alerts
}

// Delete alerts by a unique field such as "_id", or "msg_id"
//
//	// delete alerts where ids (slice) contains id
//	filter := bson.M{"_id": bson.M{"$in": ids}}
func (c *Collection) DeleteAlertsByFilter(filter bson.M) int64 {
	res, err := c.col.DeleteMany(c.Context, filter)
	if err != nil {
		panic(err)
	}

	return res.DeletedCount
}
