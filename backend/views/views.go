package views

import (
	"fmt"
	"online_code_platform_server/sqlc/database"
	"online_code_platform_server/storage"
	"os"
	"path/filepath"
)

var PathStaticFiles string

func init() {
	path, _ := os.Getwd()
	PathStaticFiles = filepath.Join(path, "..", "dist/")
}

type Post struct {
	PostId       int
	Username     string
	CodeSnipet   string
	LanguageCode string
	Comment      string
	Date         string
}

type PostTree struct {
	OriginalPost  Post
	AnswersPost   []Post
	CodeLanguages []storage.CodingLanguage
}

func (p *Post) New(row database.GetPostsFromRootRow) {
	//
	// WARNING: Username not filled properly for now, change it later when 'users' table operational
	//
	lang, _ := storage.GetLanguageDetailsFromID(int(row.LanguageID))

	p.PostId = int(row.PostID)
	p.Username = "PLACEHOLDER_NAME"
	p.CodeSnipet = row.Code
	p.LanguageCode = lang.Code
	p.Comment = row.Comment
	p.Date = row.PostDate
}

func CreateDictionaryFuncTemplate(v ...interface{}) map[string]interface{} {
	dict := map[string]interface{}{}
	lenv := len(v)

	for i := 0; i < lenv; i += 2 {
		key := fmt.Sprintf("%v", v[i])

		if i+1 >= lenv {
			dict[key] = ""
			continue
		}

		val := v[i+1]
		dict[key] = val

	}

	return dict
}
