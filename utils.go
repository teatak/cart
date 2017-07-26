package cart

type Next func()
type ComposeHandle func(*Context, Next) Next

