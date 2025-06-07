package client

import (
	"net/url"
)

type MQAUTH struct {
	User string
	Pass string
	Host string
	Port string
}

func ParseMQURL(mqURL string) (*MQAUTH, error) {
	// Parse a URL usando net/url
	u, err := url.Parse(mqURL)
	if err != nil {
		return nil, err
	}
	user := ""
	pass := ""
	if u.User != nil {
		user = u.User.Username()
		pass, _ = u.User.Password()
	}

	// Extrair host e porta
	host := u.Hostname()
	portStr := u.Port()

	return &MQAUTH{
		User: user,
		Pass: pass,
		Host: host,
		Port: portStr,
	}, nil
}
