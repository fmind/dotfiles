package <slug>

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"io/fs"
	"log/slog"
	"net/http"
	"time"

	"github.com/a-h/templ"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"<import_path>/config"
	"<import_path>/templates"
)

//go:embed static
var staticFS embed.FS

// NewAppHandler returns the root application handler.
func NewAppHandler(logger *slog.Logger, env config.Environment) http.Handler {
	// Walk assets to calculate cache-busting hashes
	initAssetHashes()

	mux := http.NewServeMux()

	// Serve embedded static files directly
	mux.Handle("/static/", http.FileServer(http.FS(staticFS)))

	// Demo user model to show everything-in-the-same-file paradigm
	demoUser := templates.User{
		ID:    "usr_01j4z3w7n5",
		Name:  "Ada Lovelace",
		Email: "ada@example.com",
	}

	// Render layout with home component using Go 1.22+ routing constraints
	mux.Handle("GET /{$}", templ.Handler(templates.Layout("GOTH App", templates.Home("Ada", demoUser))))

	// HTMX endpoint
	mux.HandleFunc("GET /api/hello", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond) // Simulate latency
		w.Header().Set("Content-Type", "text/html")
		if _, err := w.Write([]byte(`<span class="font-semibold text-teal-400">Hello from the Go backend!</span> Request processed at ` + time.Now().Format("15:04:05.000"))); err != nil {
			slog.ErrorContext(r.Context(), "write failed", "error", err)
		}
	})

	// Wrap the middleware chain in otelhttp so every request opens a server span
	// (the outermost layer, so downstream logs carry its trace/span IDs).
	return otelhttp.NewHandler(Chain(mux, SecurityHeaders(env), RequestLogger(logger)), "http.server")
}

// initAssetHashes walks the embedded static assets and populates templates.AssetHashes
// with content hashes for cache-busting.
func initAssetHashes() {
	err := fs.WalkDir(staticFS, "static", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		data, err := fs.ReadFile(staticFS, path)
		if err != nil {
			return err
		}
		h := sha256.Sum256(data)
		// Map "/static/..." path to short hash (e.g., "/static/css/dist.css" -> "abcdef01")
		templates.AssetHashes["/"+path] = hex.EncodeToString(h[:4])
		return nil
	})
	if err != nil {
		slog.Error("failed to generate static asset hashes", "error", err)
	}
}
