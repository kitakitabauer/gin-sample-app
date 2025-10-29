package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kitakitabauer/gin-sample-app/docs"
)

type DocsHandler struct{}

func NewDocsHandler() *DocsHandler {
	return &DocsHandler{}
}

func (h *DocsHandler) RegisterRoutes(router *gin.Engine) {
	router.GET("/openapi.yaml", h.serveYAML)
	router.GET("/docs/swagger", h.serveSwaggerUI)
	router.GET("/docs/redoc", h.serveRedoc)
}

func (h *DocsHandler) serveYAML(c *gin.Context) {
	data, err := docs.OpenAPIFS.ReadFile(docs.OpenAPIPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load OpenAPI spec"})
		return
	}
	c.Data(http.StatusOK, "application/yaml", data)
}

func (h *DocsHandler) serveSwaggerUI(c *gin.Context) {
	html := `<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8">
    <title>Swagger UI</title>
    <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
  </head>
  <body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js" crossorigin></script>
    <script>
      window.onload = () => {
        window.ui = SwaggerUIBundle({
          url: '/openapi.yaml',
          dom_id: '#swagger-ui',
        });
      };
    </script>
  </body>
</html>`
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}

func (h *DocsHandler) serveRedoc(c *gin.Context) {
	html := `<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8">
    <title>ReDoc</title>
    <style>
      body { margin: 0; padding: 0; }
      #redoc-container { height: 100vh; }
    </style>
  </head>
  <body>
    <div id="redoc-container"></div>
    <script src="https://cdn.jsdelivr.net/npm/redoc@next/bundles/redoc.standalone.js"></script>
    <script>
      document.addEventListener('DOMContentLoaded', function () {
        if (window.Redoc) {
          window.Redoc.init('/openapi.yaml', {}, document.getElementById('redoc-container'));
        }
      });
    </script>
  </body>
</html>`
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}
