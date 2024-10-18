package endpoints

import (
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
)

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
	Group     *echo.Group
	prefix    string
	versions  []string
	frontends []string
	parent    *EchoWrapper
}

func NewEchoWrapper(e *echo.Echo) *EchoWrapper {
	return &EchoWrapper{
		Echo: e,
	}
}

func (w *EchoWrapper) AddEnv(env ...Env) {
	w.endpoints.addEnv(env...)
}

// AddFrontends は、対象とするフロントエンドを表す識別子
// (e.g. "guest", "manager", "admin)
// を追加する
func (w *EchoWrapper) AddFrontends(frontends ...string) {
	w.endpoints.addFrontends(frontends...)
}

// AddAPI は、原則として外部から直接呼ばないこと
// ただし、wrapされたEchoを直接使ってエンドポイントを生やす場合
// （EchoWrapperが対応していないメソッドを使う場合など）
// に限り、直接呼んでよい
func (w *EchoWrapper) AddAPI(path string, desc Desc, method string) {
	w.endpoints.addAPI(API{
		Name:   desc.Name,
		Path:   path + desc.query(),
		Desc:   desc.Desc,
		Method: method,
	})
}

// AddAPITyped は、原則として外部から直接呼ばないこと
// ただし、wrapされたEchoを直接使ってエンドポイントを生やす場合
// （EchoWrapperが対応していないメソッドを使う場合など）
// に限り、直接呼んでよい
func (w *EchoWrapper) AddAPITyped(path string, desc Desc, method string, req any, resp any) {
	w.endpoints.addAPI(API{
		Name:     desc.Name,
		Path:     path + desc.query(),
		Desc:     desc.Desc,
		Method:   method,
		Request:  req,
		Response: resp,
	})
}

func (w *EchoWrapper) Generate(filename string) error {
	return w.endpoints.generate(filename)
}

func (w *EchoWrapper) GenerateOpenApiJson(filename string, config OpenApiGeneratorConfig) error {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer func() {
		file.Close()
	}()

	return w.endpoints.generateOpenApiJson(file, config)
}

func (w *EchoWrapper) GenerateOpenApi(filename string, config OpenApiGeneratorConfig) error {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer func() {
		file.Close()
	}()
	return w.endpoints.generateOpenApiYaml(file, config)
}

func (w *EchoWrapper) GET(path string, h echo.HandlerFunc, desc Desc, m ...echo.MiddlewareFunc) *echo.Route {
	w.AddAPI(path, desc, "GET")
	return w.Echo.GET(path, h, m...)
}

func (w *EchoWrapper) POST(path string, h echo.HandlerFunc, desc Desc, m ...echo.MiddlewareFunc) *echo.Route {
	w.AddAPI(path, desc, "POST")
	return w.Echo.POST(path, h, m...)
}

func (w *EchoWrapper) PUT(path string, h echo.HandlerFunc, desc Desc, m ...echo.MiddlewareFunc) *echo.Route {
	w.AddAPI(path, desc, "PUT")
	return w.Echo.PUT(path, h, m...)
}

func (w *EchoWrapper) PATCH(path string, h echo.HandlerFunc, desc Desc, m ...echo.MiddlewareFunc) *echo.Route {
	w.AddAPI(path, desc, "PATCH")
	return w.Echo.PATCH(path, h, m...)
}

func (w *EchoWrapper) DELETE(path string, h echo.HandlerFunc, desc Desc, m ...echo.MiddlewareFunc) *echo.Route {
	w.AddAPI(path, desc, "DELETE")
	return w.Echo.DELETE(path, h, m...)
}

func (w *EchoWrapper) GETTyped(path string, h echo.HandlerFunc, desc Desc, resp any, m ...echo.MiddlewareFunc) *echo.Route {
	w.AddAPITyped(path, desc, "GET", nil, resp)
	return w.Echo.GET(path, h, m...)
}

func (w *EchoWrapper) POSTTyped(path string, h echo.HandlerFunc, desc Desc, req any, resp any, m ...echo.MiddlewareFunc) *echo.Route {
	w.AddAPITyped(path, desc, "POST", req, resp)
	return w.Echo.POST(path, h, m...)
}

func (w *EchoWrapper) PUTTyped(path string, h echo.HandlerFunc, desc Desc, req any, resp any, m ...echo.MiddlewareFunc) *echo.Route {
	w.AddAPITyped(path, desc, "PUT", req, resp)
	return w.Echo.PUT(path, h, m...)
}

func (w *EchoWrapper) PATCHTyped(path string, h echo.HandlerFunc, desc Desc, req any, resp any, m ...echo.MiddlewareFunc) *echo.Route {
	w.AddAPITyped(path, desc, "PATCH", req, resp)
	return w.Echo.PATCH(path, h, m...)
}

func (w *EchoWrapper) DELETETyped(path string, h echo.HandlerFunc, desc Desc, resp any, m ...echo.MiddlewareFunc) *echo.Route {
	w.AddAPITyped(path, desc, "DELETE", nil, resp)
	return w.Echo.DELETE(path, h, m...)
}

func makeHandler[Req any, Resp any](h func(ctx echo.Context, req Req) (Resp, error)) echo.HandlerFunc {
	return func(c echo.Context) error {
		var r Req
		if err := c.Bind(&r); err != nil {
			return err
		}
		if err := c.Validate(&r); err != nil {
			return err
		}

		resp, err := h(c, r)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, resp)
	}
}

func makeHandlerNoRequest[Resp any](h func(ctx echo.Context) (Resp, error)) echo.HandlerFunc {
	return func(c echo.Context) error {
		resp, err := h(c)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, resp)
	}
}

func makeHandlerNoContent[Req any](h func(ctx echo.Context, req Req) error) echo.HandlerFunc {
	return func(c echo.Context) error {
		var r Req
		if err := c.Bind(&r); err != nil {
			return err
		}
		if err := c.Validate(&r); err != nil {
			return err
		}

		if err := h(c, r); err != nil {
			return err
		}

		return c.NoContent(http.StatusNoContent)
	}
}

// handlerを受け取ってGETエンドポイントを生やす関数
// handlerの戻り値がnilではない時は、c.JSON(http.StatusOK, resp)として返す
// handlerの戻り値がnilの場合は、c.NoContent(http.StatusOK)として返す
// NOTE: Go1.20時点では、メソッドがtype parameterをもてないので関数として定義されている
func EwGET[Resp any](w *EchoWrapper, path string, h func(ctx echo.Context) (Resp, error), desc Desc, m ...echo.MiddlewareFunc) *echo.Route {
	var resp Resp
	return w.GETTyped(path, makeHandlerNoRequest(h), desc, resp, m...)
}

// 詳細についてはEwGETを見よ
// 返すべきresponseがないケースでは、EwPOSTNoContentを使うこと
func EwPOST[Req any, Resp any](w *EchoWrapper, path string, h func(ctx echo.Context, req Req) (Resp, error), desc Desc, m ...echo.MiddlewareFunc) *echo.Route {
	var req Req
	var resp Resp
	return w.POSTTyped(path, makeHandler(h), desc, req, resp, m...)
}

// 詳細についてはEwGETを見よ
func EwPOSTNoContent[Req any](w *EchoWrapper, path string, h func(ctx echo.Context, req Req) error, desc Desc, m ...echo.MiddlewareFunc) *echo.Route {
	var req Req
	return w.POSTTyped(path, makeHandlerNoContent(h), desc, req, nil, m...)
}

// 詳細についてはEwGETを見よ
// 返すべきresponseがないケースでは、EwPUTNoContentを使うこと
func EwPUT[Req any, Resp any](w *EchoWrapper, path string, h func(ctx echo.Context, req Req) (Resp, error), desc Desc, m ...echo.MiddlewareFunc) *echo.Route {
	var req Req
	var resp Resp
	return w.PUTTyped(path, makeHandler(h), desc, req, resp, m...)
}

// 詳細についてはEwGETを見よ
func EwPUTNoContent[Req any](w *EchoWrapper, path string, h func(ctx echo.Context, req Req) error, desc Desc, m ...echo.MiddlewareFunc) *echo.Route {
	var req Req
	return w.PUTTyped(path, makeHandlerNoContent(h), desc, req, nil, m...)
}

// 詳細についてはEwGETを見よ
func EwPATCH[Req any, Resp any](w *EchoWrapper, path string, h func(ctx echo.Context, req Req) (Resp, error), desc Desc, m ...echo.MiddlewareFunc) *echo.Route {
	var req Req
	var resp Resp
	return w.PATCHTyped(path, makeHandler(h), desc, req, resp, m...)
}

// 詳細についてはEwGETを見よ
func EwPATCHNoContent[Req any](w *EchoWrapper, path string, h func(ctx echo.Context, req Req) error, desc Desc, m ...echo.MiddlewareFunc) *echo.Route {
	var req Req
	return w.PATCHTyped(path, makeHandlerNoContent(h), desc, req, nil, m...)
}

// 詳細についてはEwGETを見よ
func EwDELETE[Resp any](w *EchoWrapper, path string, h func(ctx echo.Context) (Resp, error), desc Desc, m ...echo.MiddlewareFunc) *echo.Route {
	var resp Resp
	return w.DELETETyped(path, makeHandlerNoRequest(h), desc, resp, m...)
}

// 詳細についてはEwGETを見よ
func EwDELETENoContent[Req any](w *EchoWrapper, path string, h func(ctx echo.Context, req Req) error, desc Desc, m ...echo.MiddlewareFunc) *echo.Route {
	return w.DELETETyped(path, makeHandlerNoContent(h), desc, nil, m...)
}

func (w *EchoWrapper) Group(prefix string, m ...echo.MiddlewareFunc) *GroupWrapper {
	return w.GroupWithVersionsAndFrontends(prefix, nil, nil, m...)
}

func (w *EchoWrapper) GroupWithVersionsAndFrontends(
	prefix string,
	versions []string,
	frontends []string,
	m ...echo.MiddlewareFunc,
) *GroupWrapper {
	g := w.Echo.Group(prefix, m...)
	return &GroupWrapper{
		Group:     g,
		prefix:    prefix,
		versions:  versions,
		frontends: frontends,
		parent:    w,
	}
}

// AddAPI は、原則として外部から直接呼ばないこと
// ただし、wrapされた*echo.Groupを直接使ってエンドポイントを生やす場合
// （GroupWrapperが対応していないメソッドを使う場合など）
// に限り、直接呼んでよい
func (g *GroupWrapper) AddAPI(path string, desc Desc, method string) {
	g.parent.endpoints.addAPI(API{
		Name:       desc.Name,
		Path:       g.prefix + path + desc.query(),
		Desc:       desc.Desc,
		Method:     method,
		AuthSchema: desc.AuthSchema,
		Versions:   append(g.versions, desc.Versions...),
		Frontends:  append(g.frontends, desc.Frontends...),
	})
}

func (g *GroupWrapper) AddAPITyped(path string, desc Desc, method string, req any, resp any) {
	g.parent.endpoints.addAPI(API{
		Name:       desc.Name,
		Path:       g.prefix + path + desc.query(),
		Desc:       desc.Desc,
		Method:     method,
		AuthSchema: desc.AuthSchema,
		Request:    req,
		Response:   resp,
		Versions:   append(g.versions, desc.Versions...),
		Frontends:  append(g.frontends, desc.Frontends...),
	})
}

func (g *GroupWrapper) GET(path string, h echo.HandlerFunc, desc Desc, m ...echo.MiddlewareFunc) *echo.Route {
	g.AddAPI(path, desc, "GET")
	return g.Group.GET(path, h, m...)
}

func (g *GroupWrapper) POST(path string, h echo.HandlerFunc, desc Desc, m ...echo.MiddlewareFunc) *echo.Route {
	g.AddAPI(path, desc, "POST")
	return g.Group.POST(path, h, m...)
}

func (g *GroupWrapper) PUT(path string, h echo.HandlerFunc, desc Desc, m ...echo.MiddlewareFunc) *echo.Route {
	g.AddAPI(path, desc, "PUT")
	return g.Group.PUT(path, h, m...)
}

func (g *GroupWrapper) PATCH(path string, h echo.HandlerFunc, desc Desc, m ...echo.MiddlewareFunc) *echo.Route {
	g.AddAPI(path, desc, "PATCH")
	return g.Group.PATCH(path, h, m...)
}

func (g *GroupWrapper) DELETE(path string, h echo.HandlerFunc, desc Desc, m ...echo.MiddlewareFunc) *echo.Route {
	g.AddAPI(path, desc, "DELETE")
	return g.Group.DELETE(path, h, m...)
}

func (g *GroupWrapper) GETTyped(path string, h echo.HandlerFunc, desc Desc, resp any, m ...echo.MiddlewareFunc) *echo.Route {
	g.AddAPITyped(path, desc, "GET", nil, resp)
	return g.Group.GET(path, h, m...)
}

func (g *GroupWrapper) POSTTyped(path string, h echo.HandlerFunc, desc Desc, req any, resp any, m ...echo.MiddlewareFunc) *echo.Route {
	g.AddAPITyped(path, desc, "POST", req, resp)
	return g.Group.POST(path, h, m...)
}

func (g *GroupWrapper) PUTTyped(path string, h echo.HandlerFunc, desc Desc, req any, resp any, m ...echo.MiddlewareFunc) *echo.Route {
	g.AddAPITyped(path, desc, "PUT", req, resp)
	return g.Group.PUT(path, h, m...)
}

func (g *GroupWrapper) PATCHTyped(path string, h echo.HandlerFunc, desc Desc, req any, resp any, m ...echo.MiddlewareFunc) *echo.Route {
	g.AddAPITyped(path, desc, "PATCH", req, resp)
	return g.Group.PATCH(path, h, m...)
}

func (g *GroupWrapper) DELETETyped(path string, h echo.HandlerFunc, desc Desc, resp any, m ...echo.MiddlewareFunc) *echo.Route {
	g.AddAPITyped(path, desc, "DELETE", nil, resp)
	return g.Group.DELETE(path, h, m...)
}

// 詳細についてはEwGETを見よ
func GwGET[Resp any](g *GroupWrapper, path string, h func(ctx echo.Context) (Resp, error), desc Desc, m ...echo.MiddlewareFunc) *echo.Route {
	var resp Resp
	return g.GETTyped(path, makeHandlerNoRequest(h), desc, resp, m...)
}

// 詳細についてはEwGETを見よ
func GwPOST[Req any, Resp any](g *GroupWrapper, path string, h func(ctx echo.Context, req Req) (Resp, error), desc Desc, m ...echo.MiddlewareFunc) *echo.Route {
	var req Req
	var resp Resp
	return g.POSTTyped(path, makeHandler(h), desc, req, resp, m...)
}

// 詳細についてはEwGETを見よ
func GwPOSTNoContent[Req any](g *GroupWrapper, path string, h func(ctx echo.Context, req Req) error, desc Desc, m ...echo.MiddlewareFunc) *echo.Route {
	var req Req
	return g.POSTTyped(path, makeHandlerNoContent(h), desc, req, nil, m...)
}

// 詳細についてはEwGETを見よ
func GwPUT[Req any, Resp any](g *GroupWrapper, path string, h func(ctx echo.Context, req Req) (Resp, error), desc Desc, m ...echo.MiddlewareFunc) *echo.Route {
	var req Req
	var resp Resp
	return g.PUTTyped(path, makeHandler(h), desc, req, resp, m...)
}

// 詳細についてはEwGETを見よ
func GwPUTNoContent[Req any](g *GroupWrapper, path string, h func(ctx echo.Context, req Req) error, desc Desc, m ...echo.MiddlewareFunc) *echo.Route {
	var req Req
	return g.PUTTyped(path, makeHandlerNoContent(h), desc, req, nil, m...)
}

// 詳細についてはEwGETを見よ
func GwPATCH[Req any, Resp any](g *GroupWrapper, path string, h func(ctx echo.Context, req Req) (Resp, error), desc Desc, m ...echo.MiddlewareFunc) *echo.Route {
	var req Req
	var resp Resp
	return g.PATCHTyped(path, makeHandler(h), desc, req, resp, m...)
}

// 詳細についてはEwGETを見よ
func GwPATCHNoContent[Req any](g *GroupWrapper, path string, h func(ctx echo.Context, req Req) error, desc Desc, m ...echo.MiddlewareFunc) *echo.Route {
	var req Req
	return g.PATCHTyped(path, makeHandlerNoContent(h), desc, req, nil, m...)
}

// 詳細についてはEwGETを見よ
func GwDELETE[Resp any](g *GroupWrapper, path string, h func(ctx echo.Context) (Resp, error), desc Desc, m ...echo.MiddlewareFunc) *echo.Route {
	var resp Resp
	return g.DELETETyped(path, makeHandlerNoRequest(h), desc, resp, m...)
}

// 詳細についてはEwGETを見よ
func GwDELETENoContent[Req any](g *GroupWrapper, path string, h func(ctx echo.Context, req Req) error, desc Desc, m ...echo.MiddlewareFunc) *echo.Route {
	return g.DELETETyped(path, makeHandlerNoContent(h), desc, nil, m...)
}

type Desc struct {
	Name       string
	Query      string
	Desc       string
	AuthSchema AuthSchema
	Versions   []string
	Frontends  []string
}

func (d *Desc) query() string {
	if d.Query == "" {
		return ""
	}
	return "?" + d.Query
}
