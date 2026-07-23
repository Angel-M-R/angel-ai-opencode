
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

