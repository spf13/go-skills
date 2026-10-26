# Cobra: Help Grouping, Completion, and Output

```go
// Help groups (Cobra 1.6+): story, not dump.
rootCmd.AddGroup(&cobra.Group{ID: "core", Title: "Core Commands:"})
serveCmd.GroupID = "core"

// Flag + positional completion
serveCmd.RegisterFlagCompletionFunc("output", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
    return []string{"json", "yaml", "table"}, cobra.ShellCompDirectiveNoFileComp
})
buildCmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
    return listTargets(), cobra.ShellCompDirectiveNoFileComp
}
```

```bash
myapp completion bash > /etc/bash_completion.d/myapp
myapp completion zsh  > "${fpath[1]}/_myapp"
```

Inside commands, write to `cmd.OutOrStdout()` / `cmd.ErrOrStderr()`, never `fmt.Printf` — the latter bypasses `SetOut` and breaks test capture.
