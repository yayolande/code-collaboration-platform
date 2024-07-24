package views

import (
	"fmt"
	"online_code_platform_server/sqlc/database"
	"online_code_platform_server/storage"
	"os"
	"path/filepath"
)

var PathStaticFiles string
var PathToSqueletonPage string
var PathToComponentPage string
var NameSqueletonPage string

func init() {
	// NameSqueletonPage = "partial.tmpl"
	NameSqueletonPage = "partial.html"

	path, _ := os.Getwd()
	PathStaticFiles = filepath.Join(path, "..", "dist/")
	PathToSqueletonPage = filepath.Join(PathStaticFiles, NameSqueletonPage)
	PathToComponentPage = filepath.Join(PathStaticFiles, "components.tmpl")
}

type Post struct {
	PostID       int
	Username     string
	CodeSnipet   string
	LanguageCode string
	Comment      string
	Date         string
}

type PostTree struct {
	EmptyPost     Post
	OriginalPost  Post
	AnswersPost   []Post
	CodeLanguages []storage.CodingLanguage
}

func (p *Post) New(row database.GetPostsFromRootRow) {
	lang, _ := storage.GetLanguageDetailsFromID(int(row.LanguageID))

	p.PostID = int(row.PostID)
	p.Username = row.Username
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

func SetPathToStaticFiles(staticDirectory string) {
	path, _ := os.Getwd()
	PathStaticFiles = filepath.Join(path, staticDirectory)
}
