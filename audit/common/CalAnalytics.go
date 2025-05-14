package common

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type QueryRequestBody struct {
	QueryString   string `bson:"query_string" json:"queryString"`
	LanguageFacet string `bson:"language_facet,omitempty" json:"languageFacet"`
	CategoryFacet string `bson:"category_facet,omitempty" json:"categoryFacet"`
	DocsSet       string `bson:"docs_set,omitempty" json:"docsSet"`
}

type AnalyticsReport struct {
	ID                     bson.ObjectID    `bson:"_id"`
	Query                  QueryRequestBody `bson:"query"`
	CreatedDate            time.Time        `bson:"created_date"`
	QueryDurationInSeconds float64          `bson:"query_duration_in_seconds"`
	ResultsFeedback        *ResultsFeedback `bson:"results_feedback,omitempty"`
	SummaryFeedback        *SummaryFeedback `bson:"summary_feedback,omitempty"`
}

type ResultsFeedback struct {
	IsHelpful bool   `bson:"is_helpful"`
	Comment   string `bson:"comment,omitempty"`
}

type SummaryFeedback struct {
	IsHelpful bool   `bson:"is_helpful"`
	Category  string `bson:"category,omitempty"`
	Comment   string `bson:"comment,omitempty"`
}
