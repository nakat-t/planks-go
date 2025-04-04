# planks-go 設計

## モジュール名
github.com/nakat-t/planks-go

## モジュールの概要
- 標準モジュール `log/slog` を拡張するモジュール
- 環境変数に基づいてカスタマイズした slog.Logger オブジェクトを自動で構築します。
- モジュールを import するだけで slog.SetDefault() によりデフォルトロガーを更新します。

## ロガー作成のための環境変数

### LOGGER_LEVEL
ロガーのレベルを設定します。
値には [slog.Level.UnmarshalText](https://pkg.go.dev/log/slog#Level.UnmarshalText) が受付可能な文字列を指定できます。
これは slog.HandlerOptions.Level を変更します。

### LOGGER_ADD_SOURCE
この環境変数に何らかの値が設定されていると、ログが SourceKey 属性としてソースコードの位置を出力します。
これは slog.HandlerOptions.AddSource を変更します。

### LOGGER_HANDLER
ロガーの出力ハンドラを設定します。
設定可能な値は json, text, discard のいずれかです。大文字小文字は区別しません。
これは slog.JSONHandler, slog.TextHandler, slog.DiscardHandler のどれを利用するかを決定します。
デフォルトは text になります。

### LOGGER_WRITER
ロガーの出力先 io.Writer を設定します。
設定可能な値は stdout, stderr, file のいずれかです。大文字小文字は区別しません。
デフォルトは stderr になります。
file を指定すると `LOGGER_WRITER_FILE_PATH` で指定するファイルに出力します。
このときファイルは追記モード、パーミッションは 0644 で作成されます。
この挙動は `LOGGER_WRITER_FILE_*` 変数で変更可能です。

### LOGGER_WRITER_FILE_PATH
ログをファイルに出力するときのファイルパスを設定します。

### LOGGER_WRITER_FILE_NO_APPEND
ログファイルを追記モードではなく上書きモードでファイルを開きます。

### LOGGER_WRITER_FILE_PERM
ログファイルのパーミッションを設定します。

## planks-go の制御のための環境変数

### PLANKS_NO_PANIC_ON_ERROR
この変数が定義されていると、エラーのときに planks-go は panic せず処理を続行します。
デフォルトでは planks-go は環境変数の値が不正な場合（例えば `LOGGER_LEVEL=unknown` のように期待外の値が設定されているなど）で panic します。

### PLANKS_ENV_PREFIX
この変数が定義されていると、ロガー作成のための環境変数の全ては変数名に prefix と `_` がついたものが代わりに使用されます。
例えば `PLANKS_ENV_PREFIX=MYPROG` のとき、`LOGGER_LEVEL` の代わりに `MYPROG_LOGGER_LEVEL` が使われます。

## ユースケース

`planks-go/slog/init` を import するだけで、init() で slog.Default() が自動構築されます。

environment:

```
LOGGER_LEVEL=debug
LOGGER_HANDLER=json
```

```go
import "log/slog"
import _ "github.com/nakat-t/planks-go/slog/init"

func f() {
    slog.Info("message") // Output JSON Log
}
```

--------

関連の環境変数が何も定義されていないときは、planks は何もしません。

environment:

```
```

```go
import "log/slog"
import _ "github.com/nakat-t/planks-go/slog/init"

func f() {
    slog.Info("message") // slog.Default() is not modified
}
```

--------

init() 時に自動で初期化を行わせず明示的に構築したいときは planks-go/slog の Init() を使います。

environment:

```
LOGGER_LEVEL=debug
LOGGER_HANDLER=json
```

```go
import "log/slog"
import planks_slog "github.com/nakat-t/planks-go/slog"

func f() {
    planks_slog.Init() // Build logger and set to slog.Default()
    slog.Info("message")
    
    // ...or ...
    
    logger := planks_slog.Build() // Return built logger (not set to slog.Default())
    logger.Info("message")
}
```
