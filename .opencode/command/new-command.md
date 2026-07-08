---
description: Scaffoldea un subcomando nuevo de bb siguiendo el patrón del proyecto
---

Argumentos: `$ARGUMENTS` (ej: "pr decline", "repo clone").

Creá el subcomando `bb $ARGUMENTS` siguiendo el patrón de arquitectura descrito
en AGENTS.md (Options struct + inyección de dependencias, lógica en
`internal/`). Mirá un comando existente similar en `internal/cmd/` como
referencia antes de escribir nada.

1. Archivo nuevo en `internal/cmd/<grupo>/<acción>.go`: struct `<Acción>Options`,
   `NewCmd<Acción>(f *factory.Factory) *cobra.Command`, y una función
   `<acción>Run(opts)` separada con la lógica real.
2. Registralo en el padre: `internal/cmd/pr/pr.go` para el grupo `pr`, o
   `internal/cmd/root/root.go` para `auth`/`repo`/`branch`.
3. Si hace falta un endpoint nuevo, agregalo en `internal/api/<recurso>.go`
   con su método en `*Client`, más test con `httptest.Server` y fixture en
   `internal/api/testdata/`.
4. Sumá el comando nuevo a la sección "Comandos" del README con un ejemplo.

Al terminar, corré `make build && make lint && make test` y arreglá lo que
falle.
