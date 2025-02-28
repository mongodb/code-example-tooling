package types

// CodeLengthStats holds character counts for the code example lengths in a given collection.
type CodeLengthStats struct {
	Min            int `bson:"minLength"`
	Median         int `bson:"maxLength"`
	Max            int `bson:"medianLength"`
	ShortCodeCount int `bson:"shortCodeCount"`
}
