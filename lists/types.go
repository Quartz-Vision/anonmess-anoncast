package lists

type Pushable interface {
	IsEmpty() bool
	Push(val any)
	Pop() (val any)
}
