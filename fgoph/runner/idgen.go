package runner

import gonanoid "github.com/matoous/go-nanoid/v2"

//@TODO replace this?
func makeId(prefix string) (string, error) {
	id, err := gonanoid.Generate("abcdefghijklmnopqrstuvw123456789", 10)
	return prefix + id, err
}
