package progress

type Callback func(progress float64, speed int) error
