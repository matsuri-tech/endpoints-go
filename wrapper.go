package endpoints

import "github.com/labstack/echo/v4"

// EchoWrapper に対応するメソッドが存在しないEchoの機能を使いたい場合に限り、
// wrapされたEchoを直接呼んでよい。
// ただし、それによりエンドポイントを生やす場合は、
// 当該エンドポイントの情報をEchoWrapper.AddAPI()により追加すること。
type EchoWrapper struct {
	Echo      *echo.Echo
	endpoints endpoints
}

// GroupWrapper に対応するメソッドが存在しない*echo.Groupの機能を使いたい場合に限り、
// wrapされた*echo.Groupを直接呼んでよい。
// ただし、それによりエンドポイントを生やす場合は、
// 当該エンドポイントの情報をGroupWrapper.AddAPI()により追加すること。
type GroupWrapper struct {
	Group  *echo.Group
	prefix string
	parent *EchoWrapper
}

func NewEchoWrapper(e *echo.Echo) *EchoWrapper {
	return &EchoWrapper{
		Echo: e,
	}
}

func (w *EchoWrapper) AddEnv(env ...Env) {
	w.endpoints.addEnv(env...)
}

// AddTarget は、対象とするフロントエンドを表す識別子
// (e.g. "guest", "manager", "admin)
// を追加する
func (w *EchoWrapper) AddFrontends(frontends ...string) {
	w.endpoints.addFrontends(frontends...)
}

// AddAPI は、原則として外部から直接呼ばないこと
// ただし、wrapされたEchoを直接使ってエンドポイントを生やす場合
// （EchoWrapperが対応していないメソッドを使う場合など）
// に限り、直接呼んでよい
func (w *EchoWrapper) AddAPI(path string, desc Desc) {
	w.endpoints.addAPI(API{
		Name: desc.Name,
		Path: path + desc.query(),
		Desc: desc.Desc,
	})
}

func (w *EchoWrapper) Generate(filename string) error {
	return w.endpoints.generate(filename)
}

func (w *EchoWrapper) GET(path string, h echo.HandlerFunc, desc Desc, m ...echo.MiddlewareFunc) *echo.Route {
	w.AddAPI(path, desc)
	return w.Echo.GET(path, h, m...)
}

func (w *EchoWrapper) POST(path string, h echo.HandlerFunc, desc Desc, m ...echo.MiddlewareFunc) *echo.Route {
	w.AddAPI(path, desc)
	return w.Echo.POST(path, h, m...)
}

func (w *EchoWrapper) PUT(path string, h echo.HandlerFunc, desc Desc, m ...echo.MiddlewareFunc) *echo.Route {
	w.AddAPI(path, desc)
	return w.Echo.PUT(path, h, m...)
}

func (w *EchoWrapper) PATCH(path string, h echo.HandlerFunc, desc Desc, m ...echo.MiddlewareFunc) *echo.Route {
	w.AddAPI(path, desc)
	return w.Echo.PATCH(path, h, m...)
}

func (w *EchoWrapper) DELETE(path string, h echo.HandlerFunc, desc Desc, m ...echo.MiddlewareFunc) *echo.Route {
	w.AddAPI(path, desc)
	return w.Echo.DELETE(path, h, m...)
}

func (w *EchoWrapper) Group(prefix string, m ...echo.MiddlewareFunc) *GroupWrapper {
	g := w.Echo.Group(prefix, m...)
	return &GroupWrapper{
		Group:  g,
		prefix: prefix,
		parent: w,
	}
}

// AddAPI は、原則として外部から直接呼ばないこと
// ただし、wrapされた*echo.Groupを直接使ってエンドポイントを生やす場合
// （GroupWrapperが対応していないメソッドを使う場合など）
// に限り、直接呼んでよい
func (g *GroupWrapper) AddAPI(path string, desc Desc) {
	g.parent.endpoints.addAPI(API{
		Name: desc.Name,
		Path: g.prefix + path + desc.query(),
		Desc: desc.Desc,
	})
}

func (g *GroupWrapper) GET(path string, h echo.HandlerFunc, desc Desc, m ...echo.MiddlewareFunc) *echo.Route {
	g.AddAPI(path, desc)
	return g.Group.GET(path, h, m...)
}

func (g *GroupWrapper) POST(path string, h echo.HandlerFunc, desc Desc, m ...echo.MiddlewareFunc) *echo.Route {
	g.AddAPI(path, desc)
	return g.Group.POST(path, h, m...)
}

func (g *GroupWrapper) PUT(path string, h echo.HandlerFunc, desc Desc, m ...echo.MiddlewareFunc) *echo.Route {
	g.AddAPI(path, desc)
	return g.Group.PUT(path, h, m...)
}

func (g *GroupWrapper) PATCH(path string, h echo.HandlerFunc, desc Desc, m ...echo.MiddlewareFunc) *echo.Route {
	g.AddAPI(path, desc)
	return g.Group.PATCH(path, h, m...)
}

func (g *GroupWrapper) DELETE(path string, h echo.HandlerFunc, desc Desc, m ...echo.MiddlewareFunc) *echo.Route {
	g.AddAPI(path, desc)
	return g.Group.DELETE(path, h, m...)
}

type Desc struct {
	Name  string
	Query string
	Desc  string
}

func (d *Desc) query() string {
	if d.Query == "" {
		return ""
	}
	return "?" + d.Query
}

