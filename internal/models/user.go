package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID                  primitive.ObjectID `bson:"_id,omitempty"           json:"id"`
	Email               string             `bson:"email"                   json:"email"`
	Phone               string             `bson:"phone"                   json:"phone"`
	Password            string             `bson:"password"                json:"-"`
	RefreshTokenHash    string             `bson:"refreshTokenHash,omitempty" json:"-"`
	RefreshTokenExpires *time.Time         `bson:"refreshTokenExpires,omitempty" json:"-"`
	
	// Profile completion fields
	ProfileCompletion   bool               `bson:"profileCompletion"      json:"profileCompletion"`
	FullName            string             `bson:"fullName,omitempty"     json:"fullName,omitempty"`
	Country             string             `bson:"country,omitempty"      json:"country,omitempty"`
	Address             string             `bson:"address,omitempty"      json:"address,omitempty"`
	NID                 string             `bson:"nid,omitempty"          json:"nid,omitempty"`
	PlanningMonthToStart string            `bson:"planningMonthToStart,omitempty" json:"planningMonthToStart,omitempty"`
	PlanningYearToStart  string            `bson:"planningYearToStart,omitempty"  json:"planningYearToStart,omitempty"`
	
	// Education fields
	SSC                 *Education         `bson:"ssc,omitempty"          json:"ssc,omitempty"`
	HSC                 *Education         `bson:"hsc,omitempty"          json:"hsc,omitempty"`
	HigherEducation     *HigherEducation   `bson:"higherEducation,omitempty" json:"higherEducation,omitempty"`
	
	// Language tests
	LanguageTests       []LanguageTest     `bson:"languageTests,omitempty" json:"languageTests,omitempty"`
	
	// Document fields
	SSCertificate       *Document          `bson:"ssCertificate,omitempty" json:"ssCertificate,omitempty"`
	SSCMarksheet        *Document          `bson:"sscMarksheet,omitempty" json:"sscMarksheet,omitempty"`
	
	CreatedAt           time.Time          `bson:"createdAt"               json:"createdAt"`
	UpdatedAt           time.Time          `bson:"updatedAt"               json:"updatedAt"`
}

type Education struct {
	InstitutionName string  `bson:"institutionName" json:"institutionName"`
	SchoolName      string  `bson:"schoolName"      json:"schoolName"`
	Background      string  `bson:"background"      json:"background"` // science, commerce, arts
	GPA             float64 `bson:"gpa"             json:"gpa"`
}

type HigherEducation struct {
	InstitutionName string  `bson:"institutionName" json:"institutionName"`
	SchoolName      string  `bson:"schoolName"      json:"schoolName"`
	Department      string  `bson:"department"      json:"department"`
	CGPA            float64 `bson:"cgpa"            json:"cgpa"`
}

type LanguageTest struct {
	TestType  string `bson:"testType"  json:"testType"`  // IELTS, TOEFL, etc.
	Score     string `bson:"score"     json:"score"`
	TestYear  string `bson:"testYear"  json:"testYear"`
}

type Document struct {
	FileID   string `bson:"fileId"   json:"fileId"`
	FileName string `bson:"fileName" json:"fileName"`
	FileURL  string `bson:"fileUrl"  json:"fileUrl"`
	FileSize int64  `bson:"fileSize" json:"fileSize"`
	MimeType string `bson:"mimeType" json:"mimeType"`
}
