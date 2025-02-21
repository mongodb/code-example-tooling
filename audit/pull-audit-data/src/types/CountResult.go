package types

type CountResult struct {
	ID            string `bson:"_id"`
	Count         int    `bson:"count"`
	LongCodeCount int    `bson:"longCodeCount"`
}
