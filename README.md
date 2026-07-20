# Angel AI OpenCode

```
    _                     _       _     _
    / \   _ __   __ _  ___| |     / \   | |
   / _ \ | '_ \ / _` |/ _ \ |    / _ \  | |
  / ___ \| | | | (_| |  __/ |   / ___ \ | |
  /_/   \_\_| |_|\__, |\___|_|  /_/   \_\|_|
                  |___/
```

Instalador personal de mi configuración de [opencode](https://opencode.ai): una
TUI por pasos donde selecciono qué agentes, skills, plugins y ajustes quiero
instalar en `~/.config/opencode`.

Se llama "opencode" y no solo "Angel AI" porque este repo es la configuración
para ese harness en concreto — un equivalente para otra herramienta sería un
repo hermano, no una rama de este.

## Uso

```sh
go run .                  # abre el wizard
go run . --all            # instala todo sin TUI
go run . --all --dry-run  # muestra el plan sin tocar nada
go run . --target /ruta   # instalar en otro directorio (para probar)
```

Antes de tocar `opencode.json` o `tui.json` siempre se guarda una copia
`.bak-<fecha>`, y el merge nunca borra claves existentes.

## Qué edita

Todo el contenido vive en `assets/` y se edita a mano — el código Go no hay
que tocarlo para cambiar contenido:

| Carpeta | Qué es | Se instala en |
|---|---|---|
| `assets/agents/*.md` | Un agente por archivo: frontmatter YAML + system prompt como cuerpo | `~/.config/opencode/agents/` |
| `assets/commands/*.md` | Comandos slash | `~/.config/opencode/commands/` |
| `assets/skills/<nombre>/` | Skills (carpeta completa) | `~/.config/opencode/skills/` |
| `assets/plugins/*` | Plugins JS/TS | `~/.config/opencode/plugins/` |
| `assets/themes/*.json` | Temas | `~/.config/opencode/themes/` |
| `assets/agents-md/AGENTS.md` | Reglas globales de comportamiento | `~/.config/opencode/AGENTS.md` |
| `assets/fragments/*.json` | Trozos de `opencode.json` (MCP, permisos, ajustes) que se mergean sobre el existente | `~/.config/opencode/opencode.json` |
| `assets/tui-plugins/*` | Plugins de la TUI de opencode (logo, etc.) | vía los 3 toggles del paso final del wizard, no escaneo directo |

Por ejemplo, para cambiar el system prompt del orquestador: editar
`assets/agents/angel-orchestrator.md` y volver a ejecutar el instalador. Para
añadir un agente nuevo, crear `assets/agents/mi-agente.md` — el wizard lo
detecta solo.

## Prerequisitos en la máquina destino

- [opencode](https://opencode.ai)
- CLI de OpenSpec: `npm i -g @fission-ai/openspec` (los skills `openspec-*` lo invocan)
