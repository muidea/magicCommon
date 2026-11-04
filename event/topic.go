package event

import (
	"strings"
	"unicode/utf8"

	cd "github.com/muidea/magicCommon/def"
)

const maxTopicBytes = 65535

// TopicFilter is an immutable, validated MQTT topic filter. Its zero value
// matches nothing. Address routing and HTTP routes use separate contracts.
type TopicFilter struct {
	raw    string
	levels []string
}

func validateTopicString(value, kind string) *cd.Error {
	reason := ""
	switch {
	case value == "":
		reason = "empty"
	case len(value) > maxTopicBytes:
		reason = "too long"
	case !utf8.ValidString(value):
		reason = "invalid UTF-8"
	case strings.IndexByte(value, 0) >= 0:
		reason = "contains NUL"
	}
	if reason != "" {
		return cd.NewError(cd.IllegalParam, "invalid topic "+kind+": "+reason)
	}
	return nil
}

// ValidateTopicName checks MQTT 3.1.1/5.0 common topic name requirements.
func ValidateTopicName(name string) *cd.Error {
	if err := validateTopicString(name, "name"); err != nil {
		return err
	}
	if strings.ContainsAny(name, "+#") {
		return cd.NewError(cd.IllegalParam, "invalid topic name: contains wildcard")
	}
	return nil
}

// ValidateTopicFilter checks an ordinary MQTT filter, without shared groups.
func ValidateTopicFilter(filter string) *cd.Error {
	_, err := ParseTopicFilter(filter)
	return err
}

// ParseTopicFilter validates and compiles a filter, preserving empty levels.
func ParseTopicFilter(raw string) (TopicFilter, *cd.Error) {
	if err := validateTopicString(raw, "filter"); err != nil {
		return TopicFilter{}, err
	}
	levels := strings.Split(raw, "/")
	for index, level := range levels {
		if strings.ContainsAny(level, "+#") && level != "+" && (level != "#" || index != len(levels)-1) {
			return TopicFilter{}, cd.NewError(cd.IllegalParam, "invalid topic filter: misplaced wildcard")
		}
	}
	return TopicFilter{raw: raw, levels: levels}, nil
}

// MatchTopic matches a concrete valid topic; invalid topics always return false.
func (f TopicFilter) MatchTopic(topic string) bool {
	return ValidateTopicName(topic) == nil && f.matchTopic(topic)
}

// matchTopic is the shared allocation-free matcher for already validated names.
func (f TopicFilter) matchTopic(topic string) bool {
	if f.raw == "" || (topic[0] == '$' && (f.raw[0] == '+' || f.raw[0] == '#')) {
		return false
	}
	if f.raw == topic {
		return true
	}
	for index, level := range f.levels {
		if level == "#" {
			return true
		}
		next := strings.IndexByte(topic, '/')
		value := topic
		if next >= 0 {
			value = topic[:next]
		}
		if level != "+" && level != value {
			return false
		}
		if next < 0 {
			return index == len(f.levels)-1 || (index == len(f.levels)-2 && f.levels[index+1] == "#")
		}
		topic = topic[next+1:]
	}
	return false
}

// MatchValue is a one-way Filter -> Topic Name match. It never matches addresses.
func MatchValue(pattern, val string) bool {
	filter, err := ParseTopicFilter(pattern)
	return err == nil && filter.MatchTopic(val)
}

// moreSpecific defines SimpleObserver's deterministic single-handler selection.
func (f TopicFilter) moreSpecific(other TopicFilter) bool {
	if other.raw == "" {
		return true
	}
	fExact := !strings.ContainsAny(f.raw, "+#")
	oExact := !strings.ContainsAny(other.raw, "+#")
	if fExact != oExact {
		return fExact
	}
	for index := 0; index < len(f.levels) && index < len(other.levels); index++ {
		left, right := topicLevelRank(f.levels[index]), topicLevelRank(other.levels[index])
		if left != right {
			return left > right
		}
	}
	if len(f.levels) != len(other.levels) {
		return len(f.levels) > len(other.levels)
	}
	return f.raw < other.raw
}

func topicLevelRank(level string) int {
	switch level {
	case "#":
		return 0
	case "+":
		return 1
	default:
		return 2
	}
}
