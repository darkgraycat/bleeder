package ir

type Program[T any] struct {
	commands []Command[T]
	ctx      Context
}

type Command[T any] interface {
	Execute(ctx *Context) (T, error)
}

type Context struct {
	SampleRate int
}
