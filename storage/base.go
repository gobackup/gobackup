package storage

type Base interface {
	Perform() error
}

func New(model string) (ctx Base) {
	switch model {
	case "local":
		ctx = newLocal()
	}

	return
}

func compress() {

}
