package handlers

import (
	// embed is used to include the openapi.yaml file in the binary.
	_ "embed"
	"net/http"
)

//go:embed openapi.yaml
var openapiSpec []byte

// RegisterSwagger registers the /openapi.yaml and /docs endpoints.
func RegisterSwagger(mux *http.ServeMux) {
	mux.HandleFunc("/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/yaml")
		w.Write(openapiSpec)
	})
	mux.HandleFunc("/docs", func(w http.ResponseWriter, r *http.Request) {
		html := `<!DOCTYPE html><html><head><title>Swagger UI</title><link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist/swagger-ui.css"></head><body><div id="swagger-ui"></div><script src="https://unpkg.com/swagger-ui-dist/swagger-ui-bundle.js"></script><script>window.onload=function(){SwaggerUIBundle({url:'/openapi.yaml',dom_id:'#swagger-ui'})}</script></body></html>`
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(html))
	})
}
