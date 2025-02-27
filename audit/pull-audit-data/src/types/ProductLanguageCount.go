package types

type ProductLanguageCount struct {
	ID struct {
		Product  string `bson:"product"`
		Language string `bson:"language"`
	} `bson:"_id"`
	TotalSum int `bson:"totalSum"`
}
