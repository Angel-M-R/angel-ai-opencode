# Angel AI OpenCode

Instalador personal de mi configuración de [opencode](https://opencode.ai): una TUI
por pasos donde selecciono qué agentes, comandos, skills, plugins y ajustes quiero
instalar en `~/.config/opencode`.

Se llama así (y no solo "Angel AI") porque este repo es específicamente la
configuración para el harness **opencode** — si en el futuro monto algo
equivalente para otra herramienta, será otro repo hermano.

## Uso

```sh
go run .                # abre el wizard
go run . --all          # instala todo sin TUI
go run . --all --dry-run  # muestra el plan sin tocar nada
go run . --target /ruta   # instalar en otro directorio (para probar)
```

Antes de tocar `opencode.json` siempre se guarda una copia `opencode.json.bak-<fecha>`.

## Dónde se edita cada cosa

Todo el contenido vive en `assets/` y se edita a mano — el código Go nunca hay que tocarlo
para cambiar contenido:

| Carpeta | Qué es | Se instala en |
|---|---|---|
| `assets/agents/*.md` | Un agente por archivo: frontmatter YAML (mode, tools, permission) + **system prompt como cuerpo del markdown** | `~/.config/opencode/agents/` |
| `assets/commands/*.md` | Comandos slash | `~/.config/opencode/commands/` |
| `assets/skills/<nombre>/` | Skills (carpeta completa) | `~/.config/opencode/skills/` |
| `assets/plugins/*` | Plugins JS/TS | `~/.config/opencode/plugins/` |
| `assets/themes/*.json` | Temas | `~/.config/opencode/themes/` |
| `assets/fragments/*.json` | Trozos de `opencode.json` (MCP, permisos, ajustes) que se mergean sobre el existente | `~/.config/opencode/opencode.json` |

Por ejemplo, para cambiar el system prompt del orquestador: editar
`assets/agents/angel-orchestrator.md` y volver a ejecutar el instalador.

Para añadir un agente nuevo basta con crear `assets/agents/mi-agente.md`; el wizard
lo detecta solo (el catálogo se construye escaneando `assets/`).

### Extras de la TUI (paso final, no forman parte del catálogo escaneado)

`assets/tui-plugins/` guarda `angel-logo.tsx` (logo + estado de MCP en el
footer) y su dependencia obligatoria `mcp-footer-state.ts`. A diferencia del
resto de `assets/`, estos 3 toggles no se detectan escaneando carpetas —
están fijados en `internal/install/extras.go` porque cada uno implica una
acción compuesta (copiar varios archivos y/o mergear en `tui.json`, nunca en
`opencode.json`):

- **Logo Angel AI** — copia `angel-logo.tsx` + `mcp-footer-state.ts` y añade
  la ruta del logo a `tui.json` → `plugin`.
- **Tema one-dark-pro** — activa `"theme": "one-dark-pro"` en `tui.json`.
- **Subagent statusline** — añade el plugin de terceros
  `opencode-subagent-statusline` (npm) a `tui.json` → `plugin`.

Para añadir un toggle nuevo hay que tocar Go (`ExtraOptions` +
`ApplyExtras`/`PlanExtras` en `extras.go`) — es la única parte de la
instalación donde el contenido no basta por sí solo.

## Estructura del código

- `main.go` — flags y arranque
- `internal/catalog` — escanea `assets/` y construye los items instalables
- `internal/install` — copia archivos, deep-merge de fragments en
  `opencode.json`, y los 3 extras de TUI (`extras.go`) mergeados en `tui.json`
- `internal/tui` — wizard Bubble Tea (categorías → extras → confirmación → resultado)

## Arquitectura de agentes (fase planificación/spec/apply)

- `angel-orchestrator` (primario) — coordinador delgado: entrevista, delega y rutea
  por `openspec status --json`. Gate obligatorio de entrevista antes de planificar:
  pregunta si quieres preguntas de producto + técnicas o solo técnicas, y ejecuta
  los skills `product-grilling` / `technical-grilling` en el hilo principal.
- `openspec-planner` — explora código y escribe SOLO dentro de `openspec/`
  (propose, continue, ff, update, sync, archive vía skills oficiales).
- `openspec-implementer` — implementa lotes de tasks con `openspec-apply-change`;
  tests en verde o el lote no es `done`.
- `openspec-verifier` — solo lectura + ejecución de tests; "verificado" exige
  comandos ejecutados con exit codes, nunca solo lectura de artefactos.
- Los skills `openspec-*` son los oficiales de [OpenSpec](https://github.com/Fission-AI/OpenSpec)
  v1.6.0, vendorizados sin modificar. Para actualizarlos: regenerar con
  `openspec init --tools opencode` en un stage y copiar aquí.

## Review (fase post-spec, antes de archivar)

Tras verificar el cambio, el orquestador pregunta (checkboxes) qué reviews
lanzar. Cada lente es de solo lectura, evalúa solo lo que el diff toca
(triaje), y nunca aplica cambios — solo reporta hallazgos numerados. Los fixes
solo se delegan a `openspec-implementer` sobre los hallazgos que el usuario
seleccione explícitamente.

- `review-security-risk` — ¿nos pueden atacar? ¿se puede perder/corromper
  datos? (secretos, authz, inyección, XSS, dependencias, integridad de datos).
- `review-simplicity` — ¿sobra código, abstracción, comentarios o tests?
  ¿reinventamos algo que ya existía? (anti-slop: overengineering, duplicación,
  código muerto, comentarios que narran o mienten, tests redundantes/de más).
- `review-correctness` — ¿hace lo que debe, incluso en los bordes? (lógica,
  edge cases, manejo de errores, invariantes de tipo, performance con
  evidencia).

Judgment Day (doble juicio ciego dual-blind sobre uno de estos lentes, con
fix quirúrgico) quedó diseñado pero **sin implementar por ahora** — se
retoma más adelante.

## Skills, comandos y plugins — auditoría 2026-07-20

Se revisó comando/skill/plugin por pieza y se partió de cero en vez de podar:
solo entra lo que se decide explícitamente, no lo que se hereda por defecto.

- **Eliminados por ser huérfanos** (dependían de `gentle-ai`/binarios que no
  instalamos, o eran de otra herramienta ajena colada al importar
  `~/.config/opencode` el primer día): plugins `cmux-feed.js`,
  `cmux-session.js`, `model-variants.ts`, `review-result-artifacts.ts`,
  `skill-registry.ts`; skill+comando `skill-registry` (su función —que un
  subagente sepa qué skills puede usar— ya la cubre nativamente la tool
  `skill` de opencode, sin ficheros de índice); `_shared/` (protocolo de
  inyección de rutas de skill que nuestro orquestador delgado no usa).
- **Eliminados por ser gobernanza del propio repo de gentle-ai, no genéricos**:
  `branch-pr`, `issue-creation` (reglas como "toda PR debe enlazar un issue
  aprobado" son su proceso de contribución, no un flujo reutilizable).
- **Revisados uno a uno y descartados por elección**: `chained-pr`,
  `work-unit-commits`, `skill-creator`, `skill-improver`, `comment-writer`,
  `go-testing`.
- **Entra**: `cognitive-doc-design` (guía para escribir docs con poca carga
  cognitiva).
- **Eliminado**: plugin `engram.ts` y el servidor MCP `engram` de
  `fragments/mcp.json` (2026-07-20) — sin gestor de memoria por ahora.
  Pendiente: comparativa de gestores de memoria para decidir si vuelve algo
  y con qué herramienta.

## `AGENTS.md` global — reescrito, no importado

`assets/agents-md/AGENTS.md` → `~/.config/opencode/AGENTS.md`. Se revisó
bloque a bloque en vez de importar el de gentle-ai tal cual:

- **Importado verbatim**: la guía de CodeGraph — no es contenido de gentle,
  es la misma regla que ya vive en `~/.claude/CLAUDE.md` (ya escrita por mí
  para regir cualquier herramienta). Único cambio: `gentle-ai codegraph init`
  → `codegraph init` (comando nativo, verificado con `codegraph --help`).
- **Mantenido, sin marca**: reglas genéricas de comportamiento (no firmar
  commits como IA, respuestas cortas, una pregunta cada vez, verificar antes
  de afirmar), la separación artefactos-en-inglés-siempre, y la comprobación
  obligatoria de skills disponibles antes de responder. Guía de Context7
  (mecánica, sin dependencias).
- **Eliminado por completo**: la persona "Senior Architect, GDE & MVP" de
  Gentleman, su tono (mayúsculas, corrección "ruthless"), su filosofía de
  enseñanza y su stack personal (LazyVim/Tmux/Zellij) — angel-orchestrator no
  tiene persona, solo las reglas de comportamiento de arriba.
- **Eliminado por redundante**: el bloque de "responde en el idioma del
  usuario" — es el comportamiento por defecto del modelo, no hace falta
  pedirlo; con eso se cae también el default de español rioplatense/voseo de
  gentle.
- **En pausa**: el protocolo de Engram (la mitad del archivo original) —
  depende de la comparativa de memoria pendiente.

## Prerequisitos en la máquina destino

- [opencode](https://opencode.ai)
- CLI de OpenSpec: `npm i -g @fission-ai/openspec` (los skills `openspec-*` lo invocan)

## Notas

- El contenido inicial se importó de mi `~/.config/opencode` real el 2026-07-19,
  renombrando `gentle-orchestrator` → `angel-orchestrator`. El 2026-07-19 se
  sustituyó el sistema SDD de Gentle (10 agentes + 12 comandos + skills) por el
  flujo OpenSpec oficial con 3 workers.
- El merge no borra nada: si `gentle-orchestrator` sigue definido en tu `opencode.json`,
  quedará ahí hasta que lo quites a mano (el installer solo añade/actualiza claves).
