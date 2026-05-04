package empleadas

import (
	"encoding/json"
	"salon-app/db"

	"github.com/valyala/fasthttp"
)

type Empleada struct {
	ID       int    `json:"id"`
	Nombre   string `json:"nombre"`
	Telefono string `json:"telefono"`
	Email    string `json:"email"`
	Activa   bool   `json:"activa"`
}

func ListarEmpleadas(ctx *fasthttp.RequestCtx) {
	rows, err := db.DB.Query("SELECT id, nombre, telefono, email, activa FROM empleadas WHERE activa = 1")
	if err != nil {
		ctx.SetStatusCode(500)
		ctx.SetBodyString("Error al obtener empleadas")
		return
	}
	defer rows.Close()

	var lista []Empleada
	for rows.Next() {
		var e Empleada
		rows.Scan(&e.ID, &e.Nombre, &e.Telefono, &e.Email, &e.Activa)
		lista = append(lista, e)
	}

	ctx.SetContentType("application/json")
	json.NewEncoder(ctx).Encode(lista)
}

func CrearEmpleada(ctx *fasthttp.RequestCtx) {
	var e Empleada
	if err := json.Unmarshal(ctx.PostBody(), &e); err != nil {
		ctx.SetStatusCode(400)
		ctx.SetBodyString("Datos inválidos")
		return
	}

	_, err := db.DB.Exec(
		"INSERT INTO empleadas (nombre, telefono, email) VALUES (?, ?, ?)",
		e.Nombre, e.Telefono, e.Email,
	)
	if err != nil {
		ctx.SetStatusCode(500)
		ctx.SetBodyString("Error al crear empleada")
		return
	}

	ctx.SetStatusCode(201)
	ctx.SetBodyString("Empleada creada")
}

func EliminarEmpleada(ctx *fasthttp.RequestCtx) {
	id := ctx.UserValue("id").(string)
	_, err := db.DB.Exec("UPDATE empleadas SET activa = 0 WHERE id = ?", id)
	if err != nil {
		ctx.SetStatusCode(500)
		ctx.SetBodyString("Error al eliminar empleada")
		return
	}
	ctx.SetBodyString("Empleada eliminada")
}
