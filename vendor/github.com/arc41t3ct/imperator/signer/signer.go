package signer

import (
	"fmt"
	"strings"
	"time"

	goalone "github.com/bwmarrin/go-alone"
)

type Signer struct {
	Secret []byte
}

func (s *Signer) GenerateTokenFromString(data string) string {
	var stringToSign string
	crypt := goalone.New(s.Secret, goalone.Timestamp)
	if strings.Contains(data, "?") {
		stringToSign = fmt.Sprintf("%s&hash=", data)
	} else {
		stringToSign = fmt.Sprintf("%s?hash=", data)
	}

	tokenBytes := crypt.Sign([]byte(stringToSign))
	token := string(tokenBytes)
	return token
}

func (s *Signer) VerifyToken(token string) bool {
	crypt := goalone.New(s.Secret, goalone.Timestamp)
	_, err := crypt.Unsign([]byte(token))
	if err != nil {
		return false
	}
	return true
}

func (s *Signer) Expired(token string, minutesUntilExpire int) bool {
	crypt := goalone.New(s.Secret, goalone.Timestamp)
	ts := crypt.Parse([]byte(token))
	return time.Since(ts.Timestamp) > time.Duration(minutesUntilExpire)*time.Minute
}
