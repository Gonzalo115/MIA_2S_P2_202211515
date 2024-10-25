package main

import (
	analyzer "BACKEND/analyzer"
	"fmt"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {

	//execute -path=/home/fernando/Documentos/Repositorios/Archivos/MIA_2S_P1_202211515/BACKEND/script.bash

	// Crear una nueva instancia de Fiber
	app := fiber.New()

	// Configurar el middleware CORS
	app.Use(cors.New())

	// Definir la ruta POST para recibir el comando del usuario
	app.Post("/analyze", func(c *fiber.Ctx) error {
		// Estructura para recibir el JSON
		type Request struct {
			Command string `json:"command"`
		}

		// Crear una instancia de Request
		var req Request

		// Parsear el cuerpo de la solicitud como JSON
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid JSON",
			})
		}

		// Obtener el comando del cuerpo de la solicitud
		input := req.Command
		fmt.Println("input: ", input)

		// Separar el comando en líneas
		lines := strings.Split(input, "\n")

		// Lista para acumular los resultados de salida
		var results []string

		// Analizar cada línea
		for _, line := range lines {
			// Ignorar líneas vacías
			if strings.TrimSpace(line) == "" {
				results = append(results, line)
				continue
			}

			// Si la línea comienza con '#', añadirla directamente a los resultados
			if strings.HasPrefix(strings.TrimSpace(line), "#") {
				results = append(results, line)
				continue
			}

			// Llamar a la función Analyzer del paquete analyzer para analizar la línea
			result, err := analyzer.Analyzer(line)
			if err != nil {
				// Si hay un error, almacenar el mensaje de error en lugar del resultado
				result = fmt.Sprintf("Error: %s", err.Error())
			}

			// Acumular los resultados
			results = append(results, result)
		}

		// Devolver una respuesta JSON con la lista de resultados
		return c.JSON(fiber.Map{
			"results": results,
		})
	})

	// Iniciar el servidor en el puerto 3000
	log.Fatal(app.Listen(":3000"))
}
