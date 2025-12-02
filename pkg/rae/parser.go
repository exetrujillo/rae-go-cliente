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

// FetchWordResponse representa la respuesta completa de una palabra
type FetchWordResponse struct {
	ID          string       `json:"id"`
	Header      string       `json:"encabezado"`
	Etymology   string       `json:"etimologia,omitempty"`
	Definitions []Definition `json:"definiciones"`
}

// ParseHTMLDefinitions extrae definiciones del artículo HTML
func ParseHTMLDefinitions(html string) ([]byte, error) {
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

	response := FetchWordResponse{
		ID:          id,
		Header:      header,
		Etymology:   etymology,
		Definitions: definitions,
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
