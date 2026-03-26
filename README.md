# endpoints-go

- .endpoints.json ファイルを自動生成するための Echo ラッパー

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
    Desc: "ユーザ一覧を取得する"
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

## スキーマのカスタマイズ

`invopop/jsonschema` は Go の型構造を元にスキーマを生成するため、
`MarshalJSON` によって JSON 上の型が Go の型と異なる場合（例: `uint` ベースの型が JSON では `string` になる）、
生成されるスキーマが実際の挙動と一致しないことがあります。

`WithSchemaOverride` を使うと、型ごとに1回設定するだけで
その型が使われる全フィールドに自動的に適用されます。

```go
// CurrencyType は内部的に uint だが MarshalJSON により JSON 上は string として扱われる
type CurrencyType uint

ew := endpoints.NewEchoWrapper(e,
    endpoints.WithSchemaOverride(CurrencyType(0), &jsonschema.Schema{Type: "string"}),
)
```

フィールドごとに `jsonschema:"type=string"` タグを書く必要はありません。
