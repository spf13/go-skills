# Viper: JSON Strictness and Writing Configuration Back

## JSON Config Files Are Decoded Strictly (Go 1.27)

Since Go 1.27, `encoding/json` v2 semantics apply: duplicate keys and invalid
UTF-8 in JSON strings are rejected. Viper reads JSON through `mapstructure`
which inherits this strictness when the module's floor is Go 1.27+. A config file
with `{"a":1,"a":2}` now fails where it once silently took the last value.
Guard with `viper.SetConfigType` or a build-tag-gated permissive path if you
must tolerate legacy files.

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
