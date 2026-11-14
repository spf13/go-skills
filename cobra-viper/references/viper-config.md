# Viper: JSON Config Files and Writing Configuration Back

## JSON Config Files Under Go 1.27's `encoding/json`

Since Go 1.27, `encoding/json` runs on the `encoding/json/v2` implementation,
and Viper decodes `.json` config files with `encoding/json`. The v1 API keeps v1
semantics on the new engine, so `{"a":1,"a":2}` still takes the last value and
invalid UTF-8 is still replaced with U+FFFD. What can change is error *text*:
never match a JSON decode error by message, match by type.

Strict decoding belongs to the `encoding/json/v2` package. To make a duplicate
key fail at startup, validate the raw bytes with it before `ReadConfig`:

```go
import jsonv2 "encoding/json/v2"

data, err := os.ReadFile(path)
if err != nil {
    return err
}
var probe any
if err := jsonv2.Unmarshal(data, &probe); err != nil {
    return fmt.Errorf("config %s: %w", path, err)
}
v.SetConfigType("json")
if err := v.ReadConfig(bytes.NewReader(data)); err != nil {
    return err
}
```

`GOEXPERIMENT=nojsonv2` restores the old implementation process-wide and removes
`encoding/json/v2`, so the check above stops compiling under it. `v.Unmarshal`
is unaffected: it maps fields with `mapstructure` tags, not `encoding/json`.

`v.Debug()` prints the merged settings map — use it when a key is not where you expect.

## Writing Configuration Back

Viper can serialize its merged state:

| Method | Target | Existing file |
|--------|--------|---------------|
| `v.WriteConfig()` | discovered/named path | overwritten |
| `v.SafeWriteConfig()` | `configPaths[0]`+`configName.configType` | `ConfigFileAlreadyExistsError` |
| `v.WriteConfigAs(path)` | given path | overwritten |
| `v.SafeWriteConfigAs(path)` | given path | `ConfigFileAlreadyExistsError` |

**First-run pitfall:** `WriteConfig`/`SafeWriteConfig` need a file Viper found or was named. On first run there is none, so they fail with `ConfigFileNotFoundError`. Resolve the path yourself and use `SafeWriteConfigAs`.

```go
func writeConfig(v *viper.Viper, path string) error {
    v.SetConfigType("yaml")
    if err := v.SafeWriteConfigAs(path); err != nil {
        if _, exists := errors.AsType[viper.ConfigFileAlreadyExistsError](err); exists {
            return nil
        }
        return fmt.Errorf("writing config to %s: %w", path, err)
    }
    return nil
}
```

Never write back on a read path (`PersistentPreRunE`, plain `serve` run) — a flag typo should not rewrite the user's file.

Two limits: re-encode loses comments/order/formatting; only learned keys (defaults, file, `Set`, bound flags/env) are persisted — `AutomaticEnv` unregistered keys are invisible.
