package server

import (
	"regexp"
)

func replaceWildcards(s string) string {
	return regexp.MustCompile(`\\\.\\\*`).ReplaceAllString(s, ".*")
}

func RegexpString(text string) (*regexp.Regexp, error) {
	regexPattern := "^" + regexp.QuoteMeta(text) + "$"
	regexPattern = replaceWildcards(regexPattern)
	return regexp.Compile(regexPattern)

}
func (mq *MQ) handlePub(data MQData) {

	for topic, _ := range mq.subs {
		re, err := RegexpString(topic)
		if err != nil {
			continue
		}
		if re.MatchString(data.Topic) {
			for _, sub := range mq.subs[topic] {
				mq.Send(sub, MQData{
					Cmd:      "PUB",
					Topic:    data.Topic,
					Regtopic: topic,
					Payload:  data.Payload,
				})

			}

		}

	}

}
