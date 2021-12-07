package utils

import "regexp"

func GetWordsFrom(text string) []string {
	words := regexp.MustCompile("[\\p{Hebrew}]+")
	return words.FindAllString(text, -1)
}
