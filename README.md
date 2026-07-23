
```
     _                     _       _     _
    / \   _ __   __ _  ___| |     / \   | |
   / _ \ | '_ \ / _` |/ _ \ |    / _ \  | |
  / ___ \| | | | (_| |  __/ |   / ___ \ | |
 /_/   \_\_| |_|\__, |\___|_|  /_/   \_\|_|
                 |___/
```

## Instalación

La distribución inicial admite únicamente **macOS en Apple Silicon**
(`Darwin/arm64`). No requiere Go ni clonar este repositorio. Instala la última
versión estable con:

```sh
curl --proto '=https' --tlsv1.2 -fsSL https://raw.githubusercontent.com/Angel-M-R/angel-ai-opencode/main/install.sh | /bin/sh
```

El instalador verifica la descarga y coloca el ejecutable en
`~/.local/bin/angel-ai`.

```sh
angel-ai                       # abre el wizard interactivo
angel-ai version               # muestra la versión instalada sin usar la red
angel-ai update                # fuerza una comprobación de actualización
```

## Comparativa de diseño

Existen muchos harnesses para agentes de código. Comparo solo estos tres porque,
a día de hoy, están entre mis favoritos. No es un benchmark ni pretende declarar
un ganador: resume las diferencias que más pesan en mi experiencia actual.

| Gentle AI | Angel AI | Oh My Pi |
|:---:|:---:|:---:|
| <img src="docs/images/gentle-ai-logo.png" width="56" alt="Logo de Gentle AI"> | <img src="docs/images/angel-ai-badge.svg" width="56" alt="Logo AAI de Angel AI"> | <img src="docs/images/oh-my-pi-icon.svg" width="56" alt="Icono de Oh My Pi"> |

### Agentes

| Gentle AI | Angel AI | Oh My Pi |
|---|---|---|
| 18 subagentes + 1 orquestador | 6 subagentes + 1 orquestador | 6 agentes + `advisor` |

En mi experiencia actual, los modelos más capaces, una buena compactación y el
mejor aprovechamiento del contexto reducen la necesidad de dividir el trabajo
entre muchos agentes. Más especialización puede ayudar, pero también añade
coordinación y contexto duplicado.

### Planificación y entrevistas

| Gentle AI | Angel AI | Oh My Pi |
|---|---|---|
| Sin entrevista integrada | Entrevista opcional, técnica y/o de producto | Sin entrevista integrada |

Hoy prefiero poder entrevistar antes de planificar cuando el problema todavía
está borroso, sin imponer ese paso a tareas directas. Angel AI toma ideas de
[`grill-me`](https://github.com/mattpocock/skills) y
[`gstack`](https://github.com/garrytan/gstack) para cuestionar requisitos de
producto y decisiones técnicas antes de escribir el plan.

### Memoria

| Gentle AI | Angel AI | Oh My Pi |
|---|---|---|
| Engram | Sin memoria integrada | Hindsight |

Sigo probando alternativas. En la práctica, muchas soluciones de memoria se
parecen a guardar notas; si el contexto conservado es pobre, incompleto o ya no
es válido, puede desviar al agente y resultar peor que no tener memoria.

### Specs

| Gentle AI | Angel AI | Oh My Pi |
|---|---|---|
| SDD propio con Engram, OpenSpec o ambos | OpenSpec oficial | Sistema propio |

Llevo tiempo trabajando con specs y mi criterio actual es dividir PRDs, ADRs y
features en tareas pequeñas y verificables. Elijo OpenSpec para el flujo
estructurado, aunque skills como
[`/prototype`](https://github.com/mattpocock/skills/tree/main/skills/engineering/prototype)
o [`/to-spec`](https://github.com/mattpocock/skills/tree/main/skills/engineering/to-spec)
pueden producir una spec Markdown directa cuando basta con algo más ligero.

### Ahorro de tokens

| Gentle AI | Angel AI | Oh My Pi |
|---|---|---|
| Sin optimizador específico | Sin optimizador específico | Hashline: −61 % de tokens de salida con Grok 4 Fast |

Valoro reducir reintentos y resultados inútiles, pero soy escéptico ante los
ahorradores que prometen recortes drásticos simplemente delegando en un modelo
más barato: gastar menos tokens no compensa perder precisión o repetir trabajo.

### Review final

| Gentle AI | Angel AI | Oh My Pi |
|---|---|---|
| Obligatoria; 1 lente o 4R según el riesgo | Opcional; hasta 3 revisores elegidos | Opcional; entre 1 y 16 revisores según el cambio |

Mi preferencia actual es que la revisión sea proporcional al riesgo. Gentle AI
la convierte en una puerta obligatoria y reserva sus cuatro revisores 4R para
cambios sensibles o grandes; Angel AI deja elegir hasta tres perspectivas; Oh
My Pi ajusta automáticamente el paralelismo de `/review` al tamaño del diff.

### MCPs

| Gentle AI | Angel AI | Oh My Pi |
|---|---|---|
| 3 | 2 | 4 |

Para mí, el número de MCPs no mide por sí solo la calidad del harness. Prefiero
un conjunto pequeño y deliberado: cada servidor debe aportar una capacidad útil
sin inflar el catálogo de herramientas ni distraer al modelo.


## Uso desde el repositorio

```sh
go run .                  # abre el wizard
go run . --all            # instala todo sin TUI
go run . --all --dry-run  # muestra el plan sin tocar nada
go run . --target /ruta   # instalar en otro directorio (para probar)
```


## Qué edita

Todo el contenido vive en `assets/` y se edita a mano — el código Go no hay
que tocarlo para cambiar contenido. Esto es lo que se escribe en la máquina
destino, agrupado por dónde acaba:

| Se instala en | Qué es |
|---|---|
| `~/.config/opencode/agents/` | Agentes (un archivo por agente: frontmatter YAML + system prompt) |
| `~/.config/opencode/commands/` | Comandos slash |
| `~/.config/opencode/skills/` | Skills (conserva archivos extra del destino) |
| `~/.config/opencode/plugins/` | Plugins JS/TS |
| `~/.config/opencode/themes/` | Temas |
| `~/.config/opencode/tui-plugins/` | Plugins de la TUI (logo, etc.), activados vía los 3 toggles del paso final del wizard |
| `~/.config/opencode/AGENTS.md` | Reglas globales de comportamiento (+ reglas de CodeGraph si se selecciona) |
| `~/.config/opencode/opencode.json` | MCP, permisos y ajustes que se mergean sobre el existente (+ config de CodeGraph si se selecciona) |


## Prerequisitos en la máquina destino

- [opencode](https://opencode.ai)
- `npm`, necesario si se selecciona CodeGraph y el instalador debe instalar su CLI
- CLI de OpenSpec: `npm i -g @fission-ai/openspec` (los skills `openspec-*` lo invocan)
