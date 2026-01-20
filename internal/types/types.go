package types

type App struct {
	Name string // domain (without .test) e.g. "blog" -> blog.test
	Path string // absolute path
	Port int    // assigned port
}
