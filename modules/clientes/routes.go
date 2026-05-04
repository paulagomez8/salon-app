package clientes

import "github.com/fasthttp/router"

func RegisterRoutes(r *router.Router) {
	r.GET("/clientes", ListarClientes)
	r.POST("/clientes", CrearCliente)
	r.DELETE("/clientes/{id}", EliminarCliente)
}
