package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"regexp"
	"strings"
)

// Estructura para representar un token
type Token struct {
	Lexema    string `json:"lexema"`
	Categoria string `json:"categoria"`
	Posicion  int    `json:"posicion"`
}

func main() {
	// Manejar las solicitudes HTTP
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			// Servir la página HTML
			tmpl, err := template.New("index").Parse(htmlTemplate)
			if err != nil {
				http.Error(w, fmt.Sprintf(`{"error": "%s"}`, "Error al analizar la plantilla HTML"), http.StatusInternalServerError)
				return
			}
			tmpl.Execute(w, nil)
		} else if r.Method == http.MethodPost {
			// Leer el cuerpo de la solicitud JSON
			decoder := json.NewDecoder(r.Body)
			var data map[string]string
			err := decoder.Decode(&data)
			if err != nil {
				http.Error(w, fmt.Sprintf(`{"error": "%s"}`, "Error al leer el cuerpo de la solicitud JSON"), http.StatusBadRequest)
				return
			}

			// Obtener el código Prolog del cuerpo de la solicitud
			codigoProlog, ok := data["codigoProlog"]
			if !ok {
				http.Error(w, `{"error": "Parámetro 'codigoProlog' no encontrado en la solicitud"}`, http.StatusBadRequest)
				return
			}

			// Realizar análisis léxico en Prolog
			tokens, err := AnalizarLexicoProlog(codigoProlog)
			if err != nil {
				http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err), http.StatusInternalServerError)
				return
			}

			// Convertir los resultados a formato JSON y enviarlos como respuesta
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(tokens)
		}
	})

	// Iniciar el servidor en el puerto 8080
	fmt.Println("Servidor escuchando en http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

// Función para analizar léxicamente el código fuente en Prolog
func AnalizarLexicoProlog(codigoFuente string) ([]Token, error) {
	// Definir expresiones regulares para Prolog
	expresionNumeros := regexp.MustCompile(`\d+`)
	expresionOperadores := regexp.MustCompile(`[+\-*/]`)
	expresionAsignacion := regexp.MustCompile(`=`)
	expresionParentesis := regexp.MustCompile(`[()]`)
	expresionPuntuacion := regexp.MustCompile(`[.,]`)
	expresionAtomos := regexp.MustCompile(`[a-z][a-zA-Z0-9_]*`)
	expresionHechos := regexp.MustCompile(`\w+\([\w,]+\)\s*\.`)
	expresionReglas := regexp.MustCompile(`\w+\([\w,]+\) :- [\w(),]+\.`)
	expresionConsultas := regexp.MustCompile(`\?- [\w(),]+\.`)
	expresionVariablesConsulta := regexp.MustCompile(`\b[A-Z][a-zA-Z0-9_]*\b`)
	expresionListas := regexp.MustCompile(`\[[\w,]+\]`)
	expresionOperadoresLogicos := regexp.MustCompile(`(,|;|\-> <)`)

	// Palabras reservadas de Prolog
	palabrasReservadas := map[string]string{
		"if":      "PalabraReservada",
		"else":    "PalabraReservada",
		"then":    "PalabraReservada",
		"true":    "PalabraReservada",
		"false":   "PalabraReservada",
		"cut":     "PalabraReservada",
		"fail":    "PalabraReservada",
		"not":     "PalabraReservada",
		"consult": "PalabraReservada",
		"assert":  "PalabraReservada",
		"retract": "PalabraReservada",
		"listing": "PalabraReservada",
		"write":   "PalabraReservada",
		"read":    "PalabraReservada",
		"writef":  "PalabraReservada",
		"nl":      "PalabraReservada",
		// Agrega más palabras reservadas según sea necesario...
	}

	// Expresión regular para identificar cualquier caracter no reconocido en Prolog
	expresionNoReconocido := regexp.MustCompile(`[^A-Za-z0-9_(),;=+\-*/.\s]`)

	// Dividir el código fuente en líneas y eliminar espacios en blanco adicionales
	lineas := strings.Split(codigoFuente, "\n")

	// Inicializar la lista de tokens
	var tokens []Token

	// Iterar sobre cada línea y buscar tokens en Prolog
	for i, linea := range lineas {
		// Eliminar espacios en blanco adicionales
		linea = strings.TrimSpace(linea)

		// Verificar cadena de caracteres sin cerrar
		if strings.Count(linea, `"`)%2 != 0 {
			return nil, fmt.Errorf(`Error: Cadena de caracteres sin cerrar en la línea %d`, i+1)
		}

		// Buscar hechos en la línea
		hechos := expresionHechos.FindAllString(linea, -1)
		for _, lexema := range hechos {
			tokens = append(tokens, Token{lexema, "Hecho", i + 1})
		}

		// Buscar reglas en la línea
		reglas := expresionReglas.FindAllString(linea, -1)
		for _, lexema := range reglas {
			tokens = append(tokens, Token{lexema, "Regla", i + 1})
		}

		// Buscar consultas en la línea
		consultas := expresionConsultas.FindAllString(linea, -1)
		for _, lexema := range consultas {
			tokens = append(tokens, Token{lexema, "Consulta", i + 1})
		}

		// Buscar variables en la línea de consultas
		variablesConsulta := expresionVariablesConsulta.FindAllString(linea, -1)
		for _, lexema := range variablesConsulta {
			tokens = append(tokens, Token{lexema, "Variable", i + 1})
		}

		// Buscar listas en la línea
		listas := expresionListas.FindAllString(linea, -1)
		for _, lexema := range listas {
			tokens = append(tokens, Token{lexema, "Lista", i + 1})
		}

		// Buscar operadores lógicos en la línea
		operadoresLogicos := expresionOperadoresLogicos.FindAllString(linea, -1)
		for _, lexema := range operadoresLogicos {
			tokens = append(tokens, Token{lexema, "OperadorLogico", i + 1})
		}

		// Buscar variables en la línea
		//variables := expresionVariables.FindAllString(linea, -1)
		//for _, lexema := range variables {
		//	tokens = append(tokens, Token{lexema, "Variable", i + 1})
		//}

		// Buscar números en la línea
		numeros := expresionNumeros.FindAllString(linea, -1)
		for _, lexema := range numeros {
			tokens = append(tokens, Token{lexema, "Numero", i + 1})
		}

		// Buscar operadores en la línea
		operadores := expresionOperadores.FindAllString(linea, -1)
		for _, lexema := range operadores {
			tokens = append(tokens, Token{lexema, "Operador", i + 1})
		}

		// Buscar operador de asignación en la línea
		if asignacion := expresionAsignacion.FindString(linea); asignacion != "" {
			tokens = append(tokens, Token{asignacion, "Asignacion", i + 1})
		}

		// Buscar paréntesis en la línea
		parentesis := expresionParentesis.FindAllString(linea, -1)
		for _, lexema := range parentesis {
			tokens = append(tokens, Token{lexema, "Parentesis", i + 1})
		}

		// Buscar puntuación en la línea
		puntuacion := expresionPuntuacion.FindAllString(linea, -1)
		for _, lexema := range puntuacion {
			tokens = append(tokens, Token{lexema, "Puntuacion", i + 1})
		}

		// Buscar átomos en la línea
		atomos := expresionAtomos.FindAllString(linea, -1)
		for _, lexema := range atomos {
			tokens = append(tokens, Token{lexema, "Atomo", i + 1})
		}

		// Buscar palabras reservadas en la línea
		for palabra, categoria := range palabrasReservadas {
			expresionPalabraReservada := regexp.MustCompile("\\b" + palabra + "\\b")
			if matches := expresionPalabraReservada.FindAllString(linea, -1); len(matches) > 0 {
				for _, lexema := range matches {
					tokens = append(tokens, Token{lexema, categoria, i + 1})
				}
			}
		}

		// Buscar cualquier caracter no reconocido en Prolog
		noReconocidos := expresionNoReconocido.FindAllString(linea, -1)
		for _, lexema := range noReconocidos {
			tokens = append(tokens, Token{lexema, "NoReconocido", i + 1})
		}

		// Agregar más bloques para otros tipos de tokens en Prolog...

		// Manejar errores de tokens no reconocidos en Prolog
		/*if len(noReconocidos) > 0 {
			return nil, fmt.Errorf(`Error: Caracter no reconocido en la línea %d: %s`, i+1, noReconocidos[0])
		}*/
	}

	return tokens, nil
}

const htmlTemplate = `
<!DOCTYPE html>
<html lang="es">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Análisis Léxico en Prolog</title>
    <style>
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            margin: 20px;
            background-color: #f8f9fa;
        }

        h1 {
            color: #007bff;
            text-align: center;
        }

        label {
            display: block;
            margin-top: 20px;
            color: #343a40;
        }

        textarea {
            width: 100%;
            height: 150px;
            font-size: 16px;
            padding: 10px;
            margin-bottom: 20px;
            border: 1px solid #ced4da;
            border-radius: 4px;
            box-sizing: border-box;
        }

        button {
            padding: 10px;
            font-size: 16px;
            cursor: pointer;
            background-color: #007bff;
            color: #fff;
            border: none;
            border-radius: 4px;
            box-sizing: border-box;
        }

        button:hover {
            background-color: #0056b3;
        }

        #resultado {
            margin-top: 20px;
            background-color: #fff;
            padding: 20px;
            border-radius: 4px;
            box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
        }

        h2 {
            color: #007bff;
        }

        p {
            margin-bottom: 5px;
        }

        .error {
            color: red;
        }
    </style>
</head>
<body>
    <div>
        <h1>Análisis Léxico en el lenguaje de programación Prolog</h1>
        
        <label for="codigoProlog">Ingresa la expresión regular:</label>
        <textarea id="codigoProlog" placeholder="Escribe tu código Prolog aquí..."></textarea>

        <button onclick="realizarAnalisis()">Realizar Análisis Léxico</button>

        <div id="resultado"></div>
    </div>

    <script>
        function realizarAnalisis() {
            const codigoProlog = document.getElementById('codigoProlog').value;

            fetch('/', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ codigoProlog }),
            })
            .then(response => response.json())
            .then(data => {
                mostrarResultado(data);
            })
            .catch(error => {
                mostrarError(error);
            });
        }

        function mostrarResultado(resultados) {
            const resultadoDiv = document.getElementById('resultado');
            resultadoDiv.innerHTML = '<h2>Resultados del Análisis Léxico:</h2>';

            resultados.forEach(token => {
                const tokenInfo = '<p>Lexema: ' + token.lexema + ', Categoría: ' + token.categoria + ', Posición: ' + token.posicion + '</p>';
                resultadoDiv.innerHTML += tokenInfo;
            });
        }

        function mostrarError(error) {
            const resultadoDiv = document.getElementById('resultado');
            resultadoDiv.innerHTML = '<h2>Error:</h2>';
            resultadoDiv.innerHTML += '<p class="error">' + error + '</p>';
        }
    </script>
</body>
</html>



`
