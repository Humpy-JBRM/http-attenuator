package data

type EVT interface {
	GetName() string
	GetType() string
	Run() EVTResult
}

type EVTImpl struct {
	evtType string
}

func (e *EVTImpl) GetType() string {
	return e.evtType
}

func (e *EVTImpl) Run() EVTResult {
	return nil
}

type EVTResult interface {
	Error() error

	// How long the test took in millis
	GetDuration() int64
}

func NewEVTResult(err error, durationMillis int64) EVTResult {
	return &EVTResultImpl{
		err:            err,
		durationMillis: durationMillis,
	}
}

type EVTResultImpl struct {
	err            error
	durationMillis int64
}

func (er *EVTResultImpl) Error() error {
	return er.err
}

func (er *EVTResultImpl) GetDuration() int64 {
	return er.durationMillis
}
