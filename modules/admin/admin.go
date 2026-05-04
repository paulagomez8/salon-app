package admin

import (
	"encoding/json"
	"log"
	"os"
	"salon-app/db"

	"github.com/valyala/fasthttp"
)

type Precio struct {
	ID        int      `json:"id"`
	Categoria string   `json:"categoria"`
	Nombre    string   `json:"nombre"`
	PrecioMin float64  `json:"precio_min"`
	PrecioMax *float64 `json:"precio_max"`
	Orden     int      `json:"orden"`
}

// ── AUTH ──

func checkAuth(ctx *fasthttp.RequestCtx) bool {
	cookie := string(ctx.Request.Header.Cookie("salon_session"))
	expected := os.Getenv("ADMIN_SECRET")
	return cookie == expected
}

func Login(ctx *fasthttp.RequestCtx) {
	var body struct {
		Usuario  string `json:"usuario"`
		Password string `json:"password"`
	}
	if err := json.Unmarshal(ctx.PostBody(), &body); err != nil {
		ctx.SetStatusCode(400)
		ctx.SetBodyString("Datos inválidos")
		return
	}

	log.Printf("LOGIN INTENTO: usuario='%s' password='%s'", body.Usuario, body.Password)

	var storedPassword string
	err := db.DB.QueryRow(
		"SELECT password FROM admin_config WHERE usuario = ? LIMIT 1",
		body.Usuario,
	).Scan(&storedPassword)

	log.Printf("LOGIN DB: storedPassword='%s' err=%v", storedPassword, err)

	if err != nil || storedPassword != body.Password {
		ctx.SetStatusCode(401)
		ctx.SetBodyString("Usuario o contraseña incorrectos")
		return
	}

	ctx.Response.Header.Set("Set-Cookie",
		"salon_session="+os.Getenv("ADMIN_SECRET")+"; Path=/; HttpOnly; SameSite=Strict")
	ctx.SetBodyString("ok")
}

// ── CONFIG (cambiar usuario/contraseña) ──

func ObtenerConfig(ctx *fasthttp.RequestCtx) {
	if !checkAuth(ctx) {
		ctx.SetStatusCode(401)
		ctx.SetBodyString("No autorizado")
		return
	}

	var usuario string
	err := db.DB.QueryRow("SELECT usuario FROM admin_config LIMIT 1").Scan(&usuario)
	if err != nil {
		ctx.SetStatusCode(500)
		ctx.SetBodyString("Error al obtener configuración")
		return
	}

	ctx.SetContentType("application/json")
	json.NewEncoder(ctx).Encode(map[string]string{"usuario": usuario})
}

func ActualizarConfig(ctx *fasthttp.RequestCtx) {
	if !checkAuth(ctx) {
		ctx.SetStatusCode(401)
		ctx.SetBodyString("No autorizado")
		return
	}

	var body struct {
		Usuario        string `json:"usuario"`
		PasswordActual string `json:"password_actual"`
		PasswordNueva  string `json:"password_nueva"`
	}
	if err := json.Unmarshal(ctx.PostBody(), &body); err != nil {
		ctx.SetStatusCode(400)
		ctx.SetBodyString("Datos inválidos")
		return
	}

	// Verificar contraseña actual
	var storedPassword string
	err := db.DB.QueryRow("SELECT password FROM admin_config LIMIT 1").Scan(&storedPassword)
	if err != nil || storedPassword != body.PasswordActual {
		ctx.SetStatusCode(401)
		ctx.SetBodyString("Contraseña actual incorrecta")
		return
	}

	if body.Usuario == "" || body.PasswordNueva == "" {
		ctx.SetStatusCode(400)
		ctx.SetBodyString("Usuario y contraseña nueva son requeridos")
		return
	}

	_, err = db.DB.Exec(
		"UPDATE admin_config SET usuario = ?, password = ? WHERE id = 1",
		body.Usuario, body.PasswordNueva,
	)
	if err != nil {
		ctx.SetStatusCode(500)
		ctx.SetBodyString("Error al actualizar configuración")
		return
	}

	ctx.SetBodyString("ok")
}

// ── PRECIOS ──

func ListarPrecios(ctx *fasthttp.RequestCtx) {
	if !checkAuth(ctx) {
		ctx.SetStatusCode(401)
		ctx.SetBodyString("No autorizado")
		return
	}

	rows, err := db.DB.Query(`
		SELECT id, categoria, nombre, precio_min, precio_max, orden
		FROM precios ORDER BY categoria, orden
	`)
	if err != nil {
		ctx.SetStatusCode(500)
		ctx.SetBodyString("Error al obtener precios")
		return
	}
	defer rows.Close()

	var lista []Precio
	for rows.Next() {
		var p Precio
		rows.Scan(&p.ID, &p.Categoria, &p.Nombre, &p.PrecioMin, &p.PrecioMax, &p.Orden)
		lista = append(lista, p)
	}

	ctx.SetContentType("application/json")
	json.NewEncoder(ctx).Encode(lista)
}

func ActualizarPrecio(ctx *fasthttp.RequestCtx) {
	if !checkAuth(ctx) {
		ctx.SetStatusCode(401)
		ctx.SetBodyString("No autorizado")
		return
	}

	id := ctx.UserValue("id").(string)

	var p Precio
	if err := json.Unmarshal(ctx.PostBody(), &p); err != nil {
		ctx.SetStatusCode(400)
		ctx.SetBodyString("Datos inválidos")
		return
	}

	_, err := db.DB.Exec(`
		UPDATE precios SET precio_min = ?, precio_max = ? WHERE id = ?`,
		p.PrecioMin, p.PrecioMax, id,
	)
	if err != nil {
		ctx.SetStatusCode(500)
		ctx.SetBodyString("Error al actualizar precio")
		return
	}

	ctx.SetBodyString("ok")
}

// ── PRECIOS PÚBLICOS (sin auth) ──

func ListarPreciosPublicos(ctx *fasthttp.RequestCtx) {
	rows, err := db.DB.Query(`
		SELECT id, categoria, nombre, precio_min, precio_max, orden
		FROM precios ORDER BY categoria, orden
	`)
	if err != nil {
		ctx.SetStatusCode(500)
		ctx.SetBodyString("Error al obtener precios")
		return
	}
	defer rows.Close()

	var lista []Precio
	for rows.Next() {
		var p Precio
		rows.Scan(&p.ID, &p.Categoria, &p.Nombre, &p.PrecioMin, &p.PrecioMax, &p.Orden)
		lista = append(lista, p)
	}

	ctx.SetContentType("application/json")
	json.NewEncoder(ctx).Encode(lista)
}

// ── PANEL ──

func PanelAdmin(ctx *fasthttp.RequestCtx) {
	if !checkAuth(ctx) {
		ctx.Redirect("/admin/login", 302)
		return
	}
	ctx.SendFile("templates/admin.html")
}

func PaginaLogin(ctx *fasthttp.RequestCtx) {
	ctx.SendFile("templates/login.html")
}

func Logout(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.Set("Set-Cookie", "session=; Path=/; Max-Age=0")
	ctx.Redirect("/admin/login", 302)
}
