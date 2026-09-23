package docs

import (
	"html/template"
	"net/http"
	"os"

	"fishing-notes-admin-api/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

const (
	swaggerDocPath = "desc/swagger.json"
	swaggerDocURL  = "/swagger/doc.json"
)

var swaggerIndexTemplate = template.Must(template.New("swagger-index").Parse(`<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Fishing Notes Admin API Docs</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
  <style>
    html, body {
      margin: 0;
      padding: 0;
      background: #f4f7fb;
    }
    .topbar {
      display: none;
    }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.onload = function() {
      window.ui = SwaggerUIBundle({
        url: "{{ .DocURL }}",
        dom_id: '#swagger-ui',
        deepLinking: true,
        docExpansion: 'list',
        persistAuthorization: true,
        defaultModelsExpandDepth: 2
      });
    };
  </script>
</body>
</html>
`))

type swaggerPageData struct {
	DocURL string
}

func RegisterRoutes(server *rest.Server, _ *svc.ServiceContext) {
	server.AddRoutes([]rest.Route{
		{
			Method:  http.MethodGet,
			Path:    "/swagger",
			Handler: swaggerUIHandler(),
		},
		{
			Method:  http.MethodGet,
			Path:    "/swagger/doc.json",
			Handler: swaggerDocHandler(),
		},
	})
}

func swaggerUIHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = swaggerIndexTemplate.Execute(w, swaggerPageData{DocURL: swaggerDocURL})
	}
}

func swaggerDocHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		content, err := os.ReadFile(swaggerDocPath)
		if err != nil {
			http.Error(w, "swagger document not found", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, _ = w.Write(content)
	}
}
