package types

type SubProductLanguageResult struct {
	ID struct {
		Product    string `bson:"product"`
		SubProduct string `bson:"subProduct"`
		Language   string `bson:"language"`
	} `bson:"_id"`
	TotalSum int `bson:"totalSum"`
}

type SubProductLanguageCount struct {
	Product    string
	SubProduct string
	Language   string
	Count      int
}
