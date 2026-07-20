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

La instalación reconcilia archivos individuales: si un agente, skill, tema o
plugin ya existe en la misma ruta, reemplaza solo ese archivo y conserva el
resto de la carpeta. Los archivos idénticos no se reescriben ni generan backup;
cuando cambia un archivo existente se crea una copia `.bak-<fecha>-<id>`.

`AGENTS.md` global es la excepción: al seleccionarlo se reemplaza completo. El
plan previo avisa de ello y muestra su ruta exacta. `opencode.json` y `tui.json`
se combinan de forma idempotente: los objetos se mezclan, los plugins se
reconcilian por identidad y versión, y las demás listas se reemplazan para no
duplicar ni alterar comandos posicionales.

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

Por ejemplo, para cambiar el system prompt del orquestador: editar
`assets/agents/angel-orchestrator.md` y volver a ejecutar el instalador. Para
añadir un agente nuevo, crear `assets/agents/mi-agente.md` — el wizard lo
detecta solo.

## Prerequisitos en la máquina destino

- [opencode](https://opencode.ai)
- `npm`, necesario si se selecciona CodeGraph y el instalador debe instalar su CLI
- CLI de OpenSpec: `npm i -g @fission-ai/openspec` (los skills `openspec-*` lo invocan)

## Integraciones opcionales

El paso **Integraciones y extras** permite activar o desactivar CodeGraph. Si se
selecciona, el instalador:

1. reutiliza `codegraph` cuando ya está disponible en `PATH` o instala
   `@colbymchenry/codegraph@latest` globalmente mediante npm;
2. registra `codegraph serve --mcp` como MCP local en `opencode.json`;
3. añade a `AGENTS.md` las reglas de uso e inicialización por proyecto.

Si CodeGraph se desmarca, no se instala el CLI y se elimina por completo
`mcp.codegraph`, además de retirar su bloque gestionado de `AGENTS.md`. Esa clave
pertenece a este toggle aunque su contenido se hubiera personalizado. Un binario
instalado previamente se conserva, ya que puede estar en uso fuera de OpenCode.
Context7 permanece configurado por separado como MCP remoto y no necesita un
binario local.
