package main

import (
	"go.mongodb.org/mongo-driver/v2/bson"
	"time"
)

type Feedback struct {
	ID          bson.ObjectID `bson:"_id"`
	Fingerprint Fingerprint   `bson:"fingerprint"`
	Viewport    Viewport      `bson:"viewport"`
	CreatedAt   time.Time     `bson:"createdAt"`
	SubmittedAt time.Time     `bson:"submittedAt"`
	Page        Page          `bson:"page"`
	User        User          `bson:"user"`
	Comment     string        `bson:"comment"`
	Category    string        `bson:"category"`
	Attachments []Attachment  `bson:"attachments"`
}
type Fingerprint struct {
	UserAgent string `bson:"userAgent"`
	IPAddress string `bson:"ipAddress"`
}
type Viewport struct {
	Width   int32   `bson:"width"`
	Height  int32   `bson:"height"`
	ScrollY float64 `bson:"scrollY"`
	ScrollX float64 `bson:"scrollX"`
}
type Page struct {
	Title        string `bson:"title"`
	Slug         string `bson:"slug"`
	URL          string `bson:"url"`
	DocsProperty string `bson:"docs_property"`
}
type User struct {
	StitchID    string `bson:"stitch_id"`
	SegmentID   string `bson:"segment_id"`
	IsAnonymous bool   `bson:"isAnonymous"`
	Email       string `bson:"email"`
}
type Attachment struct {
	Type     string   `bson:"type"`
	DataUri  string   `bson:"dataUri"`
	Viewport Viewport `bson:"viewport"`
	ETag     string   `bson:"ETag"`
	FileType string   `bson:"fileType"`
	Bucket   string   `bson:"bucket"`
	FileName string   `bson:"fileName"`
	URL      string   `bson:"url"`
}
