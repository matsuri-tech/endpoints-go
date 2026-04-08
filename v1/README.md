# endpoints-go/v1

- .endpoints.json ファイルを自動生成するための Echo v5 ラッパー
- Echo v4 を利用する場合は [v0.x (endpoints-go)](../) を使用してください

## Requirements

- Go 1.25.0 以上
- Echo v5

## Install

```bash
go get github.com/matsuri-tech/endpoints-go/v1
```

## v0.x (Echo v4) からの移行

import パスと、ハンドラ関数の `echo.Context` を `*echo.Context` に変更してください。

```diff
- import "github.com/matsuri-tech/endpoints-go"
+ import "github.com/matsuri-tech/endpoints-go/v1"

- import "github.com/labstack/echo/v4"
+ import "github.com/labstack/echo/v5"
```

```diff
- func myHandler(c echo.Context) error { ... }
+ func myHandler(c *echo.Context) error { ... }

- func myTypedHandler(c echo.Context, req MyReq) (MyResp, error) { ... }
+ func myTypedHandler(c *echo.Context, req MyReq) (MyResp, error) { ... }
```

ルートメソッドの戻り値が `*echo.Route` から `echo.RouteInfo` に変わります。戻り値を使用していない場合は変更不要です。

## Usage

```go
e := echo.New()

ew := endpoints.NewEchoWrapper(e)

// バージョンごとのドメインの設定
ew.AddEnv(
    endpoints.Env{
        Version: "v1",
        Domain: endpoints.Domain{
            Local: "http://localhost:8000",
            LocalDev: "https://local-dev.hoge.com",
            Dev: "https://dev.hoge.com",
            Prod: "https://hoge.com",
        },
    },
    endpoints.Env{
        Version: "v2",
        Domain: endpoints.Domain{
            Local: "http://localhost:8000",
            LocalDev: "https://local-dev.hoge.com",
            Dev: "https://v2.dev.hoge.com",
            Prod: "https://v2.hoge.com",
        },
    },
)

// Frontendの設定はOptionalのため、
// 何も設定しなくても動作する
ew.AddFrontends("guest", "manager")

// 以上の設定の結果、
// "v1", "guest-v1", "manager-v1",
// "v2", "guest-v2", "manager-v2"
// の6種類の.endpoints.jsonが生成される

// エンドポイントの追加
// Versionsが指定されていない場合、全てのバージョンに含まれる
// Frontendsが指定されていない場合、全てのフロントエンド向けの設定に含まれる
ew.GET("/users", userHandler.GetUsers, endpoints.Desc{
    Name: "userIndex",
    Query: "page=2&itemsPerPage=50",
    Desc: "ユーザ一覧を取得する",
})

// このエンドポイントは"v2","guest-v2","manager-v2"のみに含まれる
ew.GET("/messages", messageHandler.GetMessages, endpoints.Desc{
    Name: "messageIndex",
    Query: "page=2&itemsPerPage=50",
    Desc: "メッセージ一覧を取得する",
    Versions: endpoints.Versions{"v2"},
})

// このエンドポイントは"v1", "manager-v1", "v2", "manager-v2"のみに含まれる
ew.GET("/inquiries", inquiryHandler.GetInquiries, endpoints.Desc{
    Name: "inquiryIndex",
    Query: "",
    Desc: "問い合わせ一覧を取得する",
    Frontends: endpoints.Frontends{"manager"},
})

// このエンドポイントは"v2", "guest-v2"のみに含まれる
ew.GET("/favorites", favoriteHandler.GetFavorites, endpoints.Desc{
    Name: "favoriteIndex",
    Query: "",
    Desc: "お気に入り一覧を取得する",
    Versions: endpoints.Versions{"v2"},
    Frontends: endpoints.Frontends{"guest"},
})

// グループの作成とエンドポイントの追加
articles := ew.Group("/articles")
articles.POST("/", articleHandler.CreateArticle, endpoints.Desc{
    Name: "createArticle",
    Query: "",
    Desc: "記事を新規作成する",
})

// グループ単位でVersionsやFrontendsを指定することもできる
comments := ew.GroupWithVersionsAndFrontends(
    "/comments",
    []string{"v2"},
    []string{"manager"},
)
// このエンドポイントは"v2", "manager-v2"のみに含まれる
comments.POST("/", commentHandler.CreateComment, endpoints.Desc{
    Name: "createComment",
    Query: "",
    Desc: "コメントを新規作成する",
})

// .endpoints.jsonファイルの出力
if err := ew.Generate(".endpoints.json"); err != nil {
    log.Printf("failed to generate endpoints file: %v", err)
}
```

## 型付きハンドラ

ジェネリック関数を使うことで、リクエスト/レスポンスの型情報を OpenAPI スキーマに自動反映できます。

```go
// リクエストとレスポンスの両方がある場合
endpoints.EwPOST[CreateUserInput, CreateUserOutput](ew, "/users", func(c *echo.Context, req CreateUserInput) (CreateUserOutput, error) {
    // ...
}, endpoints.Desc{Name: "createUser", Desc: "ユーザを作成する"})

// レスポンスのみの場合
endpoints.EwGET[GetUsersOutput](ew, "/users", func(c *echo.Context) (GetUsersOutput, error) {
    // ...
}, endpoints.Desc{Name: "getUsers", Desc: "ユーザ一覧を取得する"})

// リクエストのみ（204 No Content）の場合
endpoints.EwPOSTNoContent[UpdateUserInput](ew, "/users/:id", func(c *echo.Context, req UpdateUserInput) error {
    // ...
}, endpoints.Desc{Name: "updateUser", Desc: "ユーザを更新する"})

// Group版も同様（Gw プレフィックス）
endpoints.GwPOST[CreateArticleInput, CreateArticleOutput](articles, "/", func(c *echo.Context, req CreateArticleInput) (CreateArticleOutput, error) {
    // ...
}, endpoints.Desc{Name: "createArticle", Desc: "記事を新規作成する"})
```
