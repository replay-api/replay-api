package docs

import (
	"embed"
	"html/template"
	"net/http"
	"strings"
)

//go:embed openapi.yaml
var openapiSpec embed.FS

// SwaggerConfig holds configuration for the Swagger UI
type SwaggerConfig struct {
	Title       string
	SpecURL     string
	SwaggerURL  string
	RedocURL    string
	DarkMode    bool
	PrimaryColor string
}

// DefaultSwaggerConfig returns default configuration
func DefaultSwaggerConfig() SwaggerConfig {
	return SwaggerConfig{
		Title:        "LeetGaming Pro API",
		SpecURL:      "/api/docs/openapi.yaml",
		SwaggerURL:   "/api/docs/swagger",
		RedocURL:     "/api/docs/redoc",
		DarkMode:     true,
		PrimaryColor: "#00FF87", // LeetGaming brand color
	}
}

// SwaggerUIHandler serves the Swagger UI
func SwaggerUIHandler(config SwaggerConfig) http.HandlerFunc {
	tmpl := template.Must(template.New("swagger").Parse(swaggerUITemplate))

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "public, max-age=3600")

		if err := tmpl.Execute(w, config); err != nil {
			http.Error(w, "Failed to render Swagger UI", http.StatusInternalServerError)
		}
	}
}

// RedocHandler serves the ReDoc documentation
func RedocHandler(config SwaggerConfig) http.HandlerFunc {
	tmpl := template.Must(template.New("redoc").Parse(redocTemplate))

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "public, max-age=3600")

		if err := tmpl.Execute(w, config); err != nil {
			http.Error(w, "Failed to render ReDoc", http.StatusInternalServerError)
		}
	}
}

// OpenAPISpecHandler serves the OpenAPI specification
func OpenAPISpecHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := openapiSpec.ReadFile("openapi.yaml")
		if err != nil {
			http.Error(w, "OpenAPI spec not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/x-yaml")
		w.Header().Set("Cache-Control", "public, max-age=3600")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(data)
	}
}

// DocsIndexHandler serves the documentation index page
func DocsIndexHandler(config SwaggerConfig) http.HandlerFunc {
	tmpl := template.Must(template.New("index").Parse(docsIndexTemplate))

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		if err := tmpl.Execute(w, config); err != nil {
			http.Error(w, "Failed to render docs index", http.StatusInternalServerError)
		}
	}
}

// RegisterDocsRoutes registers all documentation routes
func RegisterDocsRoutes(mux *http.ServeMux, basePath string) {
	config := DefaultSwaggerConfig()

	// Ensure base path starts with / and doesn't end with /
	basePath = "/" + strings.Trim(basePath, "/")

	mux.HandleFunc(basePath, DocsIndexHandler(config))
	mux.HandleFunc(basePath+"/swagger", SwaggerUIHandler(config))
	mux.HandleFunc(basePath+"/redoc", RedocHandler(config))
	mux.HandleFunc(basePath+"/openapi.yaml", OpenAPISpecHandler())
}

const swaggerUITemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}} - Swagger UI</title>
    <link rel="icon" type="image/png" href="https://leetgaming.gg/favicon.png">
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
    <style>
        :root {
            --leet-primary: {{.PrimaryColor}};
            --leet-dark: #0a0a0f;
            --leet-surface: #14141f;
            --leet-text: #e0e0e0;
        }

        body {
            margin: 0;
            padding: 0;
            {{if .DarkMode}}
            background: var(--leet-dark);
            {{end}}
        }

        {{if .DarkMode}}
        .swagger-ui {
            background: var(--leet-dark);
        }

        .swagger-ui .topbar {
            background: var(--leet-surface);
            border-bottom: 2px solid var(--leet-primary);
        }

        .swagger-ui .topbar .download-url-wrapper .select-label {
            color: var(--leet-text);
        }

        .swagger-ui .info .title {
            color: var(--leet-primary);
        }

        .swagger-ui .info p, .swagger-ui .info li {
            color: var(--leet-text);
        }

        .swagger-ui .opblock-tag {
            color: var(--leet-text);
            border-bottom: 1px solid var(--leet-surface);
        }

        .swagger-ui .opblock {
            background: var(--leet-surface);
            border: 1px solid #2a2a3a;
        }

        .swagger-ui .opblock .opblock-summary-method {
            background: var(--leet-primary);
            color: var(--leet-dark);
            font-weight: bold;
        }

        .swagger-ui .opblock.opblock-get .opblock-summary-method {
            background: #61affe;
        }

        .swagger-ui .opblock.opblock-post .opblock-summary-method {
            background: var(--leet-primary);
        }

        .swagger-ui .opblock.opblock-put .opblock-summary-method {
            background: #fca130;
        }

        .swagger-ui .opblock.opblock-delete .opblock-summary-method {
            background: #f93e3e;
        }

        .swagger-ui .opblock-description-wrapper p {
            color: var(--leet-text);
        }

        .swagger-ui .btn {
            color: var(--leet-text);
            border-color: var(--leet-primary);
        }

        .swagger-ui .btn:hover {
            background: var(--leet-primary);
            color: var(--leet-dark);
        }

        .swagger-ui table thead tr th {
            color: var(--leet-text);
        }

        .swagger-ui table tbody tr td {
            color: var(--leet-text);
        }

        .swagger-ui .model-box {
            background: var(--leet-surface);
        }

        .swagger-ui .model {
            color: var(--leet-text);
        }

        .swagger-ui .scheme-container {
            background: var(--leet-surface);
        }

        .swagger-ui select {
            background: var(--leet-dark);
            color: var(--leet-text);
            border: 1px solid var(--leet-primary);
        }
        {{end}}

        .custom-header {
            background: linear-gradient(90deg, var(--leet-surface) 0%, var(--leet-dark) 100%);
            padding: 20px;
            display: flex;
            align-items: center;
            gap: 20px;
            border-bottom: 2px solid var(--leet-primary);
        }

        .custom-header img {
            height: 40px;
        }

        .custom-header h1 {
            color: var(--leet-primary);
            margin: 0;
            font-family: 'Orbitron', sans-serif;
        }

        .custom-header .nav {
            margin-left: auto;
            display: flex;
            gap: 20px;
        }

        .custom-header .nav a {
            color: var(--leet-text);
            text-decoration: none;
            padding: 8px 16px;
            border: 1px solid var(--leet-primary);
            border-radius: 4px;
            transition: all 0.2s;
        }

        .custom-header .nav a:hover, .custom-header .nav a.active {
            background: var(--leet-primary);
            color: var(--leet-dark);
        }
    </style>
    <link href="https://fonts.googleapis.com/css2?family=Orbitron:wght@700&display=swap" rel="stylesheet">
</head>
<body>
    <div class="custom-header">
        <h1>🎮 {{.Title}}</h1>
        <div class="nav">
            <a href="{{.SwaggerURL}}" class="active">Swagger UI</a>
            <a href="{{.RedocURL}}">ReDoc</a>
            <a href="{{.SpecURL}}" target="_blank">OpenAPI Spec</a>
        </div>
    </div>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
    <script>
        window.onload = function() {
            SwaggerUIBundle({
                url: "{{.SpecURL}}",
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIBundle.SwaggerUIStandalonePreset
                ],
                layout: "StandaloneLayout",
                persistAuthorization: true,
                displayRequestDuration: true,
                filter: true,
                tryItOutEnabled: true
            });
        }
    </script>
</body>
</html>`

const redocTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}} - API Reference</title>
    <link rel="icon" type="image/png" href="https://leetgaming.gg/favicon.png">
    <link href="https://fonts.googleapis.com/css2?family=Orbitron:wght@700&family=Inter:wght@400;500;600&display=swap" rel="stylesheet">
    <style>
        :root {
            --leet-primary: {{.PrimaryColor}};
            --leet-dark: #0a0a0f;
            --leet-surface: #14141f;
            --leet-text: #e0e0e0;
        }

        body {
            margin: 0;
            padding: 0;
            font-family: 'Inter', sans-serif;
        }

        .custom-header {
            background: linear-gradient(90deg, var(--leet-surface) 0%, var(--leet-dark) 100%);
            padding: 20px;
            display: flex;
            align-items: center;
            gap: 20px;
            border-bottom: 2px solid var(--leet-primary);
            position: sticky;
            top: 0;
            z-index: 100;
        }

        .custom-header h1 {
            color: var(--leet-primary);
            margin: 0;
            font-family: 'Orbitron', sans-serif;
        }

        .custom-header .nav {
            margin-left: auto;
            display: flex;
            gap: 20px;
        }

        .custom-header .nav a {
            color: var(--leet-text);
            text-decoration: none;
            padding: 8px 16px;
            border: 1px solid var(--leet-primary);
            border-radius: 4px;
            transition: all 0.2s;
        }

        .custom-header .nav a:hover, .custom-header .nav a.active {
            background: var(--leet-primary);
            color: var(--leet-dark);
        }
    </style>
</head>
<body>
    <div class="custom-header">
        <h1>🎮 {{.Title}}</h1>
        <div class="nav">
            <a href="{{.SwaggerURL}}">Swagger UI</a>
            <a href="{{.RedocURL}}" class="active">ReDoc</a>
            <a href="{{.SpecURL}}" target="_blank">OpenAPI Spec</a>
        </div>
    </div>
    <redoc spec-url='{{.SpecURL}}'
        theme='{
            "colors": {
                "primary": { "main": "{{.PrimaryColor}}" },
                "success": { "main": "#00FF87" },
                {{if .DarkMode}}
                "text": { "primary": "#e0e0e0", "secondary": "#a0a0a0" },
                "border": { "dark": "#2a2a3a", "light": "#3a3a4a" }
                {{end}}
            },
            {{if .DarkMode}}
            "sidebar": {
                "backgroundColor": "#14141f",
                "textColor": "#e0e0e0"
            },
            "rightPanel": {
                "backgroundColor": "#0a0a0f"
            },
            {{end}}
            "typography": {
                "fontFamily": "Inter, sans-serif",
                "headings": { "fontFamily": "Orbitron, sans-serif" },
                "code": { "fontFamily": "JetBrains Mono, monospace" }
            }
        }'
    ></redoc>
    <script src="https://cdn.redoc.ly/redoc/latest/bundles/redoc.standalone.js"></script>
</body>
</html>`

const docsIndexTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}} - API Documentation</title>
    <link rel="icon" type="image/png" href="https://leetgaming.gg/favicon.png">
    <link href="https://fonts.googleapis.com/css2?family=Orbitron:wght@700&family=Inter:wght@400;500;600&display=swap" rel="stylesheet">
    <style>
        :root {
            --leet-primary: {{.PrimaryColor}};
            --leet-dark: #0a0a0f;
            --leet-surface: #14141f;
            --leet-text: #e0e0e0;
        }

        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            background: var(--leet-dark);
            color: var(--leet-text);
            font-family: 'Inter', sans-serif;
            min-height: 100vh;
            display: flex;
            flex-direction: column;
            align-items: center;
            justify-content: center;
            padding: 40px;
        }

        .container {
            max-width: 800px;
            text-align: center;
        }

        h1 {
            font-family: 'Orbitron', sans-serif;
            color: var(--leet-primary);
            font-size: 3rem;
            margin-bottom: 1rem;
        }

        .subtitle {
            font-size: 1.25rem;
            color: #808080;
            margin-bottom: 3rem;
        }

        .cards {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
            gap: 24px;
            margin-bottom: 3rem;
        }

        .card {
            background: var(--leet-surface);
            border: 1px solid #2a2a3a;
            border-radius: 12px;
            padding: 32px;
            text-decoration: none;
            color: inherit;
            transition: all 0.3s;
        }

        .card:hover {
            border-color: var(--leet-primary);
            transform: translateY(-4px);
            box-shadow: 0 8px 32px rgba(0, 255, 135, 0.15);
        }

        .card h2 {
            color: var(--leet-primary);
            font-family: 'Orbitron', sans-serif;
            margin-bottom: 12px;
            font-size: 1.5rem;
        }

        .card p {
            color: #808080;
            line-height: 1.6;
        }

        .card .icon {
            font-size: 3rem;
            margin-bottom: 16px;
        }

        .links {
            display: flex;
            gap: 16px;
            justify-content: center;
            flex-wrap: wrap;
        }

        .links a {
            color: var(--leet-text);
            text-decoration: none;
            padding: 12px 24px;
            border: 1px solid #2a2a3a;
            border-radius: 8px;
            transition: all 0.2s;
        }

        .links a:hover {
            border-color: var(--leet-primary);
            color: var(--leet-primary);
        }

        .version {
            margin-top: 3rem;
            color: #404040;
            font-size: 0.875rem;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🎮 {{.Title}}</h1>
        <p class="subtitle">Choose your preferred documentation format</p>

        <div class="cards">
            <a href="{{.SwaggerURL}}" class="card">
                <div class="icon">📋</div>
                <h2>Swagger UI</h2>
                <p>Interactive API explorer with try-it-out functionality. Perfect for testing endpoints and understanding request/response formats.</p>
            </a>

            <a href="{{.RedocURL}}" class="card">
                <div class="icon">📖</div>
                <h2>ReDoc</h2>
                <p>Beautiful, three-panel documentation viewer. Great for reading comprehensive API documentation and understanding the data models.</p>
            </a>

            <a href="{{.SpecURL}}" target="_blank" class="card">
                <div class="icon">📄</div>
                <h2>OpenAPI Spec</h2>
                <p>Raw OpenAPI 3.1 specification in YAML format. Use this for code generation or importing into API clients.</p>
            </a>
        </div>

        <div class="links">
            <a href="https://leetgaming.gg" target="_blank">🏠 LeetGaming Home</a>
            <a href="https://github.com/leetgaming-pro" target="_blank">⭐ GitHub</a>
            <a href="https://discord.gg/leetgaming" target="_blank">💬 Discord</a>
        </div>

        <p class="version">API Version 2.0.0 • OpenAPI 3.1</p>
    </div>
</body>
</html>`

