
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
que tocarlo para cambiar contenido:

| Carpeta | Qué es | Se instala en |
|---|---|---|
| `assets/agents/*.md` | Un agente por archivo: frontmatter YAML + system prompt como cuerpo | `~/.config/opencode/agents/` |
| `assets/commands/*.md` | Comandos slash | `~/.config/opencode/commands/` |
| `assets/skills/<nombre>/` | Archivos de cada skill; conserva archivos extra del destino | `~/.config/opencode/skills/` |
| `assets/plugins/*` | Plugins JS/TS | `~/.config/opencode/plugins/` |
| `assets/themes/*.json` | Temas | `~/.config/opencode/themes/` |
| `assets/agents-md/AGENTS.md` | Reglas globales de comportamiento | `~/.config/opencode/AGENTS.md` |
| `assets/fragments/*.json` | Trozos de `opencode.json` (MCP, permisos, ajustes) que se mergean sobre el existente | `~/.config/opencode/opencode.json` |
| `assets/integrations/codegraph/*` | Configuración MCP y reglas que solo se aplican cuando se selecciona CodeGraph | `opencode.json` y `AGENTS.md` |
| `assets/tui-plugins/*` | Plugins de la TUI de opencode (logo, etc.) | vía los 3 toggles del paso final del wizard, no escaneo directo |


## Prerequisitos en la máquina destino

- [opencode](https://opencode.ai)
- `npm`, necesario si se selecciona CodeGraph y el instalador debe instalar su CLI
- CLI de OpenSpec: `npm i -g @fission-ai/openspec` (los skills `openspec-*` lo invocan)

