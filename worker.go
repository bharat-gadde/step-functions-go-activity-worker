package sfnactivityworker

type worker interface {
	Tune(size int)
	Start()
	Stop()
	Submit(interface{}) error
}
