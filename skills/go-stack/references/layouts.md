# Go Project Layouts

Reference trees for the go-stack scaffolding workflow; every file maps to a reference in the skill.

## CLI + Library Project Layout

```text
<slug>/
├── cmd/
│   └── <slug>/
│       └── main.go         // CLI entry point
├── config/
│   └── config.go           // Typed environment configuration
├── .env
├── .env.example
├── .gitignore
├── .golangci.yml
├── dprint.json
├── lefthook.yml
├── mise.toml
├── AGENTS.md
├── LICENSE
├── <slug>.go               // Library entry point
├── <slug>_test.go          // Unit tests
└── README.md
```

## Web + Library Project Layout

```text
<slug>/
├── cmd/
│   └── <slug>/
│       └── main.go         // Daemon entry point
├── config/
│   └── config.go           // Typed environment configuration
├── .env
├── .env.example
├── .gitignore
├── .golangci.yml
├── .air.toml
├── dprint.json
├── lefthook.yml
├── mise.toml
├── AGENTS.md
├── LICENSE
├── <slug>.go               // Core business logic / client
├── <slug>_test.go          // Core business logic tests
├── server.go               // HTTP handler and asset serving definitions
├── server_test.go          // HTTP routing and integration tests
├── middleware.go           // Standard HTTP middlewares
├── telemetry.go            // OpenTelemetry setup (SetupOTel) + slog trace correlation
├── README.md
├── assets/                     // Authored sources — never embedded
│   ├── css/
│   │   └── styles.css          // Tailwind entry point
│   └── js/
│       ├── app.js              // esbuild entry point
│       └── components/
│           └── user-card.js    // One Alpine component per module
├── scripts/
│   └── vendor.go
├── static/                     // Build output + vendor — embedded via go:embed
│   ├── css/
│   │   └── dist.css            // tailwindcss --minify
│   ├── js/
│   │   └── dist.js             // esbuild --bundle --minify
│   └── vendor/
│       ├── htmx.min.js
│       └── alpine.min.js
└── templates/
    ├── home.templ
    └── layout.templ
```

