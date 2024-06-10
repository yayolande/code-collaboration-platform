package storage

import "fmt"

type CodingLanguage struct {
	ID    int
	Code  string
	Label string
}

var CodeLanguages = [...]CodingLanguage{
	{ID: 1, Code: "js", Label: "JavaScript"},
	{ID: 2, Code: "java", Label: "Java"},
	{ID: 3, Code: "c", Label: "C"},
	{ID: 4, Code: "go", Label: "Go"},
	{ID: 5, Code: "cpp", Label: "C++"},
	{ID: 6, Code: "php", Label: "PHP"},
	{ID: 7, Code: "python", Label: "Python"},
	{ID: 8, Code: "html", Label: "HTML"},
	{ID: 9, Code: "sql", Label: "SQL"},
}

func GetLanguageDetailsFromID(id int) (CodingLanguage, error) {
	var language CodingLanguage

	err := fmt.Errorf("Language not found for 'ID = %d' in coding language storage", id)

	for _, lang := range CodeLanguages {
		if lang.ID == id {
			language = lang
			err = nil
		}
	}

	return language, err
}

func GetLanguageDetailsFromCode(code string) (CodingLanguage, error) {
	var language CodingLanguage

	err := fmt.Errorf("Language not found for 'Code = %s' in coding language storage", code)

	for _, lang := range CodeLanguages {
		if lang.Code == code {
			language = lang
			err = nil
		}
	}

	return language, err
}
