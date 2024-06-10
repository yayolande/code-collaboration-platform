// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0

package database

type CodeLanguage struct {
	LanguageID int64
	Label      string
}

type Post struct {
	PostID     int64
	UserID     int64
	LanguageID int64
	Code       string
	Comment    string
	PostDate   string
}

type PostsTree struct {
	PostID       int64
	ParentPostID int64
}

type User struct {
	UserID   int64
	Username string
	Password string
	Status   int64
}