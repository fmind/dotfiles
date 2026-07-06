package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/caarlos0/env/v11"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.41.0"
	"google.golang.org/genai"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/agent/llmagent"
	"google.golang.org/adk/v2/cmd/launcher"
	"google.golang.org/adk/v2/cmd/launcher/full"
	"google.golang.org/adk/v2/model/gemini"
	"google.golang.org/adk/v2/telemetry"
	"google.golang.org/adk/v2/tool"
	"google.golang.org/adk/v2/tool/functiontool"

	"<import_path>/config"
)

// vertexConfig holds Vertex AI settings parsed from the environment. Credentials
// come from Application Default Credentials (`gcloud auth application-default login`
// locally, the attached service account in production) — no API key is stored.
// Project is required so startup fails fast when the environment is unset.
type vertexConfig struct {
	Project  string `env:"GOOGLE_CLOUD_PROJECT,required"`
	Location string `env:"GOOGLE_CLOUD_LOCATION" envDefault:"global"`
}

// greetArgs is the typed input for the greet tool. The jsonschema tags document
// each field for the model — parse tool input into trusted types at the boundary.
type greetArgs struct {
	Name string `json:"name" jsonschema:"the name of the person to greet"`
}

// greetResult is the typed output returned to the model.
type greetResult struct {
	Message string `json:"message"`
}

// greet is a self-contained example tool. Real tools call into the library
// (<slug>.go) instead of embedding business logic in the agent entry point.
func greet(_ agent.Context, in greetArgs) (greetResult, error) {
	return greetResult{Message: fmt.Sprintf("Hello, %s!", in.Name)}, nil
}

func main() {
	ctx := context.Background()

	// Load configuration and initialize structured logging (fail fast on error).
	cfg, err := config.Load()
	if err != nil {
		slog.Error("load configuration", "error", err)
		os.Exit(1)
	}
	logger := slog.New(cfg.NewHandler(os.Stderr))
	slog.SetDefault(logger)

	// Gemini on Vertex AI with Application Default Credentials (no API key). For
	// AI Studio instead, drop Backend/Project/Location and set APIKey from env.
	vx, err := env.ParseAs[vertexConfig]()
	if err != nil {
		logger.Error("load vertex configuration", "error", err)
		os.Exit(1)
	}
	model, err := gemini.NewModel(ctx, "gemini-flash-latest", &genai.ClientConfig{
		Backend:  genai.BackendVertexAI,
		Project:  vx.Project,
		Location: vx.Location,
	})
	if err != nil {
		logger.Error("create model", "error", err)
		os.Exit(1)
	}

	// Wrap a typed Go function as a tool; the generic signature infers the schema.
	greetTool, err := functiontool.New(functiontool.Config{
		Name:        "greet",
		Description: "Greet a person by name.",
	}, greet)
	if err != nil {
		logger.Error("create tool", "error", err)
		os.Exit(1)
	}

	// Name must be a valid identifier (letters, digits, underscore), unique in the
	// agent tree. Delegate to SubAgents to build multi-agent trees.
	root, err := llmagent.New(llmagent.Config{
		Name:        "<slug>",
		Model:       model,
		Description: "A helpful assistant.",
		Instruction: "You are a helpful assistant. Use the greet tool when asked to greet someone.",
		Tools:       []tool.Tool{greetTool},
	})
	if err != nil {
		logger.Error("create agent", "error", err)
		os.Exit(1)
	}

	// full.NewLauncher wires an in-memory session service and a CLI with two modes:
	// `console` (interactive) and `web` (HTTP server hosting the `webui` dev UI, the
	// `api` REST endpoint, and `a2a` sublaunchers). Run `go run ./cmd/<slug> console`
	// (no arguments defaults to console; an unrecognized mode prints the usage).
	// Attribute ADK's spans (LLM calls, tool calls) to this service. The launcher
	// owns the exporter: OTLP via OTEL_EXPORTER_OTLP_ENDPOINT, or GCP Cloud Trace
	// with `--otel_to_cloud` (credentials reuse the Vertex ADC above).
	res := resource.NewSchemaless(semconv.ServiceName("<slug>"))
	l := full.NewLauncher()
	launcherCfg := &launcher.Config{
		AgentLoader:      agent.NewSingleLoader(root),
		TelemetryOptions: []telemetry.Option{telemetry.WithResource(res)},
	}
	if err := l.Execute(ctx, launcherCfg, os.Args[1:]); err != nil {
		logger.Error("run agent", "error", err)
		fmt.Fprintln(os.Stderr, l.CommandLineSyntax())
		os.Exit(1)
	}
}
