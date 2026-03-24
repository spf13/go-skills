# Idiomatic Go: The Pragmatic Playbook

A definitive guide to building robust, efficient, and maintainable Go applications—without the enterprise baggage.

## Another Go project layout guide? No thanks.

As Go has grown in popularity, developers are importing structural baggage from other languages—specifically Java and Spring Boot—and rebranding them as "best practices."

If you search for "Go project layout," you will inevitably be pointed to repositories like the notorious `golang-standards/project-layout`. These templates promote deeply nested directories, arbitrary architectural layers (`service/`, `repository/`, `pkg/`), heavy worker pools, and complex mocking frameworks.

**Those are not Go patterns. They are Java patterns translated into Go syntax.** They actively fight against Go's design. Go was engineered for simplicity, readability, and flat, discoverable APIs. When you force a layered, Spring-Boot-style architecture onto Go, you introduce circular dependencies, obscure the control flow, and destroy the beauty of the language its breathtaking simplicity.

This problem has gotten significantly worse with LLMs. AI coding assistants have been trained on the full corpus of the internet—including all of those misguided "Go best practices" guides. The result is that they confidently generate Java-in-Go-syntax by default, then argue with you when you push back. These skills exist to correct that. They give the LLM authoritative, first-principles guidance so it stops reaching for `internal/` junk drawers, BDD frameworks, and static worker pools when writing your Go code.

## A Course Correction

As the creator of tools like Hugo, Cobra, and Viper, and having spent years on the core Go team, I feel I have a good sense of what a well structured Go codebase should look like and I'm tired of arguing with LLMs trained on Java style codebases. This playbook is to save my sanity.

It strips away the noise and focuses on the actual architectures used by the Go standard library and the most successful, high-performance open-source projects in the world.

**This playbook advocates for:**

- **Domains over Layers:** Deleting the `internal/` junk drawer and the `pkg/` anti-pattern. Organize code by what it *does*, not by what *kind of file* it is.
- **Standard Library over Frameworks:** Using `testing`, table-driven tests, and simple stubs instead of heavy BDD or mock-generation frameworks.
- **Channels over Mutexes:** Using Go's native concurrency primitives for orchestration rather than rigid, static worker pools.
- **Command-First Architecture:** Treating your application binary as a router for commands, entirely decoupled from your core business logic.

## The Playbooks

### 1. [General Go Patterns](./go/SKILL.md)

The core language idioms. This covers fundamental Go philosophy, proper interface design, standard-library-compliant testing, and concurrency patterns. **Stop writing Java in Go.** Read this first.

### 2. [CLI Architecture: Cobra & Viper](./cobra-viper/SKILL.md)

The definitive guide to building modern command-line applications. This covers the "Command-First" architecture, 12-factor configuration, and how to structure your application using Cobra and Viper exactly as they were intended to be used.

## Installation

### Claude Code (via marketplace)

Add this repo as a Claude Code marketplace, then install individual skills:

```
/plugin marketplace add mosajjal/go-skills
/plugin install go@go-skills
/plugin install cobra-viper@go-skills
```

### Other AI Agents (Copilot, etc.)

```powershell
# Windows — directory junction (run as Administrator)
New-Item -ItemType Junction -Path "$env:USERPROFILE\.agents\skills\go" -Target "$PWD\go\skills\go"
New-Item -ItemType Junction -Path "$env:USERPROFILE\.agents\skills\cobra-viper" -Target "$PWD\cobra-viper\skills\cobra-viper"
```

```bash
# macOS / Linux
ln -s "$PWD/go/skills/go"                   "$HOME/.agents/skills/go"
ln -s "$PWD/cobra-viper/skills/cobra-viper" "$HOME/.agents/skills/cobra-viper"
```

## The Golden Rule

*Clear is better than clever.* Go code should be boring in the best possible way—predictable, consistent, and immediately understandable to a new developer opening the file for the first time. When in doubt, delete the abstraction.

## License

MIT
