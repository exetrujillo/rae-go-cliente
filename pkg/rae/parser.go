package rae

import (
	"encoding/json"
	"regexp"
	"strings"
)

// Definition representa una definición de una palabra
type Definition struct {
	Type       string   `json:"tipo"`
	Definition string   `json:"definicion"`
	Synonyms   []string `json:"sinonimos,omitempty"`
	Antonyms   []string `json:"antonimos,omitempty"`
}

type Conjugation struct {
	Mode    string              `json:"modo"` // Indicativo, Subjuntivo, Imperativo, Formas no personales
	Tenses  map[string][]string `json:"tiempos"` // Presente -> [yo como, tú comes...]
}

// FetchWordResponse representa la respuesta completa de una palabra
type FetchWordResponse struct {
	ID           string        `json:"id"`
	Header       string        `json:"encabezado"`
	Etymology    string        `json:"etimologia,omitempty"`
	Definitions  []Definition  `json:"definiciones"`
	Conjugations []Conjugation `json:"conjugaciones,omitempty"`
}

// ParseHTMLDefinitions extrae definiciones del artículo HTML
func ParseHTMLDefinitions(html string, withConjugations bool) ([]byte, error) {
	// Verificar si es un artículo HTML
	if !strings.Contains(html, "<article id=") {
		// No es HTML, devolver tal cual
		return []byte(html), nil
	}

	// Extraer ID
	idRe := regexp.MustCompile(`id="(\w+)"`)
	idMatches := idRe.FindStringSubmatch(html)
	var id string
	if len(idMatches) > 1 {
		id = idMatches[1]
	}

	// Extraer encabezado (palabra)
	headerRe := regexp.MustCompile(`<header[^>]+>(.*?)(?:</i>)?</h`)
	headerMatches := headerRe.FindStringSubmatch(html)
	var header string
	if len(headerMatches) > 1 {
		header = headerMatches[1]
		// Limpiar entidades HTML y superíndices
		header = cleanHTMLEntities(header)
		header = regexp.MustCompile(`<sup>\d+</sup>`).ReplaceAllString(header, "")
	}

	// Extraer Etimología
	var etymology string
	etymologyRe := regexp.MustCompile(`(?s)<p class="n2">(.*?)</p>`)
	if etymMatches := etymologyRe.FindStringSubmatch(html); len(etymMatches) > 1 {
		etymology = etymMatches[1]
		// Limpiar etiquetas HTML de la etimología
		etymology = regexp.MustCompile(`<[^>]+>`).ReplaceAllString(etymology, "")
		etymology = cleanHTMLEntities(etymology)
		etymology = strings.TrimSpace(etymology)
	}

	// Extraer definiciones de las etiquetas <p class="j">
	var definitions []Definition
	
	// Encontrar todas las definiciones de párrafo - usar (?s) para coincidir con nuevas líneas
	paragraphRe := regexp.MustCompile(`(?s)<p class="j"[^>]*>(.*?)</p>`)
	paragraphs := paragraphRe.FindAllStringSubmatch(html, -1)

	for _, paragraph := range paragraphs {
		if len(paragraph) < 2 {
			continue
		}
		
		content := paragraph[1]
		
		// Extraer tipo (abbr con class="d" o class="g")
		typeRe := regexp.MustCompile(`<abbr[^>]*title="([^"]+)"[^>]*>`)
		typeMatches := typeRe.FindStringSubmatch(content)
		var defType string
		if len(typeMatches) > 1 {
			defType = typeMatches[1]
			defType = cleanHTMLEntities(defType)
		}

		// Extraer sinónimos en línea para esta definición específica
		var synonyms []string
		
		// Buscar tabla de sinónimos dentro de este párrafo
		// Usar (?s) para permitir coincidencias a través de nuevas líneas
		// Nota: El HTML de la RAE para sinónimos en línea a menudo está mal formado y falta </ul>, así que coincidimos hasta </td>
		synTableRe := regexp.MustCompile(`(?s)<table class='sinonimos'>.*?<ul[^>]*>(.*?)</td>`)
		if synMatches := synTableRe.FindStringSubmatch(content); len(synMatches) > 1 {
			synRe := regexp.MustCompile(`<mark[^>]*>([^<]+)</mark>`)
			synWords := synRe.FindAllStringSubmatch(synMatches[1], -1)
			for _, match := range synWords {
				if len(match) > 1 {
					word := strings.TrimSpace(match[1])
					if word != "" && !contains(synonyms, word) {
						synonyms = append(synonyms, word)
					}
				}
			}
		}

		// Extraer antónimos en línea para esta definición específica
		var antonyms []string
		
		// Buscar div/tabla de antónimos
		// Usar (?s) para permitir coincidencias a través de nuevas líneas
		antInlineRe := regexp.MustCompile(`(?s)<div class="ant-header ant-inline">.*?<ul[^>]*>(.*?)</ul>.*?</div>`)
		if antMatches := antInlineRe.FindStringSubmatch(content); len(antMatches) > 1 {
			antRe := regexp.MustCompile(`<mark[^>]*>([^<]+)</mark>`)
			antWords := antRe.FindAllStringSubmatch(antMatches[1], -1)
			for _, match := range antWords {
				if len(match) > 1 {
					word := strings.TrimSpace(match[1])
					if word != "" && !contains(antonyms, word) {
						antonyms = append(antonyms, word)
					}
				}
			}
		}

		// Eliminar todas las etiquetas HTML y limpiar la definición
		definition := content
		// Eliminar etiquetas abbr
		definition = regexp.MustCompile(`<abbr[^>]+>.*?</abbr>`).ReplaceAllString(definition, "")
		// Eliminar etiquetas span
		definition = regexp.MustCompile(`<span class="h">.*?</span>`).ReplaceAllString(definition, "")
		definition = regexp.MustCompile(`<span class="n_acep">\S+ </span>`).ReplaceAllString(definition, "")
		// Eliminar divs/tablas de sinónimos/antónimos del texto de la definición
		definition = regexp.MustCompile(`(?s)<div class="[^"]*-header[^"]*-inline">.*?</div>`).ReplaceAllString(definition, "")
		// Eliminar todas las etiquetas HTML restantes
		definition = regexp.MustCompile(`<[^>]+>`).ReplaceAllString(definition, "")
		// Limpiar entidades HTML
		definition = cleanHTMLEntities(definition)
		// Limpiar espacios en blanco
		definition = strings.TrimSpace(definition)
		
		// Reemplazar abreviaturas
		definition = strings.ReplaceAll(definition, "sing.", "singular")
		definition = strings.ReplaceAll(definition, "pl.", "plural")
		definition = strings.ReplaceAll(definition, "t.", "también")
		definition = strings.ReplaceAll(definition, "p.", "poco")

		if definition != "" {
			definitions = append(definitions, Definition{
				Type:       defType,
				Definition: definition,
				Synonyms:   synonyms,
				Antonyms:   antonyms,
			})
		}
	}

	var conjugations []Conjugation
	if withConjugations {
		// Extraer tabla de conjugación
		cnjRe := regexp.MustCompile(`(?s)<table class="cnj">(.*?)</table>`)
		cnjMatches := cnjRe.FindStringSubmatch(html)
		if len(cnjMatches) > 1 {
			tableContent := cnjMatches[1]
			
			// Dividir por filas
			rowsRe := regexp.MustCompile(`(?s)<tr[^>]*>(.*?)</tr>`)
			rows := rowsRe.FindAllStringSubmatch(tableContent, -1)
			
			var currentMode string
			var headers []string
			var currentTenses map[string][]string
			
			// Modos válidos conocidos
			validModes := map[string]bool{
				"Formas no personales": true,
				"Indicativo":           true,
				"Subjuntivo":           true,
				"Imperativo":           true,
			}
			
			// Inicializar
			currentMode = ""
			currentTenses = make(map[string][]string)
			
			for _, row := range rows {
				rowContent := row[1]
				
				// Verificar si es un header de Modo o Tense Group (colspan="2" o más, y th)
				modeRe := regexp.MustCompile(`(?s)<th[^>]*colspan="[^"]*"[^>]*>(.*?)</th>`)
				if modeMatch := modeRe.FindStringSubmatch(rowContent); len(modeMatch) > 1 {
					headerText := cleanHTMLEntities(strings.TrimSpace(modeMatch[1]))
					
					if validModes[headerText] {
						// Es un nuevo Modo
						// Guardar el modo anterior si existe y tenía datos
						if currentMode != "" && len(currentTenses) > 0 {
							conjugations = append(conjugations, Conjugation{
								Mode:   currentMode,
								Tenses: currentTenses,
							})
						}
						
						// Iniciar nuevo modo
						currentMode = headerText
						currentTenses = make(map[string][]string)
						headers = nil
					} else {
						// Es un header de Tiempo que ocupa toda la fila (ej. "Pretérito imperfecto" en Subjuntivo)
						// Lo tratamos como el único header para las siguientes filas
						headers = []string{headerText}
					}
					continue
				}
				
				// Verificar si es una fila de headers de Tiempos (th sin colspan o con colspan=1)
				if strings.Contains(rowContent, "<th") && !strings.Contains(rowContent, "colspan") {
					headerRe := regexp.MustCompile(`(?s)<th>(.*?)</th>`)
					headerMatches := headerRe.FindAllStringSubmatch(rowContent, -1)
					headers = nil
					for _, h := range headerMatches {
						val := cleanHTMLEntities(strings.TrimSpace(h[1]))
						// Filtrar headers de metadatos
						if val == "Número" || val == "Personas del discurso" || val == "Pronombres personales" {
							continue
						}
						// Manejar headers vacíos (común en Imperativo)
						if val == "" && currentMode == "Imperativo" {
							val = "Imperativo"
						}
						headers = append(headers, val)
					}
					continue
				}
				
				// Verificar si es una fila de datos (td)
				if strings.Contains(rowContent, "<td") {
					cellRe := regexp.MustCompile(`(?s)<td[^>]*>(.*?)</td>`)
					cells := cellRe.FindAllStringSubmatch(rowContent, -1)
					
					// Si no tenemos modo asignado (ej. Formas no personales al principio sin header explícito a veces),
					// intentamos deducirlo o saltamos.
					// En el HTML de RAE, "Formas no personales" suele tener header explícito.
					if currentMode == "" {
						// Fallback para la primera sección si falta header
						currentMode = "Formas no personales"
					}
					
					if len(headers) > 0 {
						// Caso 1: Headers simples (Infinitivo, Gerundio) o Tiempos compuestos
						// Si tenemos headers, intentamos mapear
						
						// Lógica para Modos Personales vs No Personales
						if currentMode == "Formas no personales" || currentMode == "Participio" {
							for i, h := range headers {
								if i < len(cells) {
									val := cleanHTMLEntities(strings.TrimSpace(regexp.MustCompile(`<[^>]+>`).ReplaceAllString(cells[i][1], "")))
									if val != "" {
										currentTenses[h] = append(currentTenses[h], val)
									}
								}
							}
						} else {
							// Modos Personales (Indicativo, Subjuntivo, Imperativo)
							// Estructura: [Persona], [Tiempo1], [Tiempo2]...
							// O si el header era único (Pretérito imperfecto), Estructura: [Persona], [Valor]
							
							// Detectar persona
							var person string
							verbCells := cells
							
							// Si hay más celdas que headers, asumimos que las primeras son persona/pronombres
							if len(cells) > len(headers) {
								// La persona suele ser la celda anterior a los verbos
								// Pero cuidado con celdas vacías de estructura
								// Simplificación: Tomamos las últimas N celdas como verbos, donde N = len(headers)
								verbCells = cells[len(cells)-len(headers):]
								
								// La persona está en alguna de las celdas anteriores. 
								// Tomamos la última de las "anteriores" que no esté vacía o sea la inmediata anterior
								personCell := cells[len(cells)-len(headers)-1][1]
								person = cleanHTMLEntities(strings.TrimSpace(regexp.MustCompile(`<[^>]+>`).ReplaceAllString(personCell, "")))
							}
							
							for i, h := range headers {
								if i < len(verbCells) {
									val := cleanHTMLEntities(strings.TrimSpace(regexp.MustCompile(`<[^>]+>`).ReplaceAllString(verbCells[i][1], "")))
									if val != "" {
										if person != "" {
											val = person + " " + val
										}
										currentTenses[h] = append(currentTenses[h], val)
									}
								}
							}
						}
					} else {
						// Sin headers definidos (caso raro en RAE si parseamos bien)
						// Podría ser el caso de Participio si no pillamos el header
						if currentMode == "Participio" || strings.Contains(currentMode, "Participio") {
							if len(cells) > 0 {
								val := cleanHTMLEntities(strings.TrimSpace(regexp.MustCompile(`<[^>]+>`).ReplaceAllString(cells[len(cells)-1][1], "")))
								currentTenses["Participio"] = append(currentTenses["Participio"], val)
							}
						}
					}
				}
			}
			
			// Agregar el último modo
			if currentMode != "" && len(currentTenses) > 0 {
				conjugations = append(conjugations, Conjugation{
					Mode:   currentMode,
					Tenses: currentTenses,
				})
			}
		}
	}

	response := FetchWordResponse{
		ID:           id,
		Header:       header,
		Etymology:    etymology,
		Definitions:  definitions,
		Conjugations: conjugations,
	}

	return json.Marshal(response)
}

func cleanHTMLEntities(s string) string {
	s = strings.ReplaceAll(s, "&#xE1;", "á")
	s = strings.ReplaceAll(s, "&#xE9;", "é")
	s = strings.ReplaceAll(s, "&#xED;", "í")
	s = strings.ReplaceAll(s, "&#xF3;", "ó")
	s = strings.ReplaceAll(s, "&#xFA;", "ú")
	s = strings.ReplaceAll(s, "&#xF1;", "ñ")
	s = strings.ReplaceAll(s, "&#x2016;", "||")
	return s
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
