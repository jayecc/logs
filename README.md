# logs
Based on slog packaging, supports context tracking and asynchronous hooks

### Example

```go

var (
    level   slog.LevelVar // info
    syncer  = &lumberjack.Logger{Filename: "logger.log", LocalTime: true, MaxAge: 1}
)

func main() {
	
    defer syncer.Close()
    
    level.Set(logs.ParseLevel("debug"))

    slog.SetDefault(slog.New(WrapHandler(slog.NewJSONHandler(
        io.MultiWriter(os.Stdout, NewWriter(syncer, 10<<10)),
        &slog.HandlerOptions{Level: &level}),
    )))

    slog.DebugContext(ctx, "DebugContext")
    slog.InfoContext(ctx, "InfoContext")
    slog.ErrorContext(ctx, "ErrorContext")
    son := slog.With(slog.String("name", "son"))
    son.ErrorContext(ctx, "son.ErrorContext")
}

```

```bash
{"time":"2024-02-07T15:00:45.5510225+08:00","level":"DEBUG","msg":"DebugContext","trace":"b012ab24f1f864214f93435c150915ab","span":"9ed99b05197f2389"}
{"time":"2024-02-07T15:00:45.5757939+08:00","level":"INFO","msg":"InfoContext","trace":"b012ab24f1f864214f93435c150915ab","span":"9ed99b05197f2389"}                   
{"time":"2024-02-07T15:00:45.5757939+08:00","level":"ERROR","msg":"ErrorContext","trace":"b012ab24f1f864214f93435c150915ab","span":"9ed99b05197f2389"}                 
{"time":"2024-02-07T15:00:45.5763476+08:00","level":"ERROR","msg":"son.ErrorContext","name":"son","trace":"b012ab24f1f864214f93435c150915ab","span":"9ed99b05197f2389"
```