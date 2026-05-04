package clientes

import (
	"encoding/json"
	"salon-app/db"

	"github.com/valyala/fasthttp"
)

type Cliente struct {
	ID       int    `json:"id"`
	Nombre   string `json:"nombre"`
	Telefono string `json:"telefono"`
	Email    string `json:"email"`
	Notas    string `json:"notas"`
}

func ListarClientes(ctx *fasthttp.RequestCtx) {
	rows, err := db.DB.Query("SELECT id, nombre, telefono, email, notas FROM clientes")
	if err != nil {
		ctx.SetStatusCode(500)
		ctx.SetBodyString("Error al obtener clientes")
		return
	}
	defer rows.Close()

	var lista []Cliente
	for rows.Next() {
		var c Cliente
		rows.Scan(&c.ID, &c.Nombre, &c.Telefono, &c.Email, &c.Notas)
		lista = append(lista, c)
	}

	ctx.SetContentType("application/json")
	json.NewEncoder(ctx).Encode(lista)
}

func CrearCliente(ctx *fasthttp.RequestCtx) {
	var c Cliente
	if err := json.Unmarshal(ctx.PostBody(), &c); err != nil {
		ctx.SetStatusCode(400)
		ctx.SetBodyString("Datos inválidos")
		return
	}

	_, err := db.DB.Exec(
		"INSERT INTO clientes (nombre, telefono, email, notas) VALUES (?, ?, ?, ?)",
		c.Nombre, c.Telefono, c.Email, c.Notas,
	)
	if err != nil {
		ctx.SetStatusCode(500)
		ctx.SetBodyString("Error al crear cliente")
		return
	}

	ctx.SetStatusCode(201)
	ctx.SetBodyString("Cliente creado")
}

func EliminarCliente(ctx *fasthttp.RequestCtx) {
	id := ctx.UserValue("id").(string)
	_, err := db.DB.Exec("DELETE FROM clientes WHERE id = ?", id)
	if err != nil {
		ctx.SetStatusCode(500)
		ctx.SetBodyString("Error al eliminar cliente")
		return
	}
	ctx.SetBodyString("Cliente eliminado")
}
