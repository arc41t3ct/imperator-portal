package imperator

import (
	"crypto/rand"
	"os"
)

const randomString = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_+"

// CreateRadomString generates a random string from the const randomString given n the length
func (i *Imperator) CreateRadomString(n int) string {
	s, r := make([]rune, n), []rune(randomString)
	for i := range s {
		p, _ := rand.Prime(rand.Reader, len(r))
		x, y := p.Uint64(), uint64(len(r))
		s[i] = r[x%y]
	}
	return string(s)
}

func (i *Imperator) CreateDirIfNotExists(path string) error {
	const mode = 0755
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.Mkdir(path, mode)
		if err != nil {
			return err
		}
	}
	return nil
}

func (i *Imperator) CreateFileIfNotExists(path string) error {
	var _, err = os.Stat(path)
	if os.IsNotExist(err) {
		var file, err = os.Create(path)
		if err != nil {
			return nil
		}

		defer func(file *os.File) {
			_ = file.Close()
		}(file)
	}
	return nil
}
