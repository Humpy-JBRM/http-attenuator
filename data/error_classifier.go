package data

import "regexp"

type ErrorClassifier interface {
	Classify(err error) string
}

type ErrorClassifierImpl struct {
	// TODO(john): encapsulate this into its own error classifier
	errorClasses map[string]*regexp.Regexp
}

func (e *ErrorClassifierImpl) Classify(err error) string {
	for ec := range e.errorClasses {
		if e.errorClasses[ec].MatchString(err.Error()) {
			return ec
		}
	}

	return ""
}

var errorClassifier ErrorClassifier

func init() {
	// TODO(john): read these from config so that new classes of errors dont need rebuilding of code
	errorClassifier = &ErrorClassifierImpl{
		errorClasses: map[string]*regexp.Regexp{
			"dns":       regexp.MustCompile("lookup .* on .*: server misbehaving"),
			"connreset": regexp.MustCompile("connection reset by peer"),
		},
	}
}

func GetErrorClassifier() ErrorClassifier {
	return errorClassifier
}
