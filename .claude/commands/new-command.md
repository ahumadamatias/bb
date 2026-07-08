---
description: Scaffoldea un subcomando nuevo de bb siguiendo el patrón del proyecto
argument-hint: "<grupo> <acción>, ej: pr decline"
---

Argumentos: `$ARGUMENTS` (ej: "pr decline", "repo clone").

Creá el subcomando `bb $ARGUMENTS` siguiendo exactamente el patrón que ya usan
los comandos existentes en `internal/cmd/` (mirá `internal/cmd/pr/comment.go` o
`internal/cmd/pr/approve.go` como referencia antes de escribir nada). Pasos:

1. **Archivo del comando**: creá `internal/cmd/<grupo>/<acción>.go` con:
   - Un struct `<Acción>Options` que incluya `IO *iostreams.IOStreams`,
     `Client func() (*api.Client, error)`, `GitContext func() (*gitctx.Context, error)`,
     `Config func() (*config.Resolved, error)` según lo que el comando necesite,
     más los campos de flags propios del comando.
   - `NewCmd<Acción>(f *factory.Factory) *cobra.Command` que arma el Options
     struct desde `f`, define flags con `cmd.Flags()`, y en `RunE` copia los
     valores de flags globales relevantes (`f.Options.Workspace`, `f.Options.Repo`,
     `f.Options.Output`) antes de llamar a la función `xxxRun`.
   - Una función `<acción>Run(opts *<Acción>Options) error` separada del
     `cobra.Command`, con toda la lógica real. Nunca escribas a `os.Stdout`
     directo — siempre `opts.IO.Out` / `opts.IO.ErrOut`.
   - Si el comando opera sobre un repo específico, resolvé workspace/repo con
     `cmdutil.ResolveWorkspaceRepo` (o `cmdutil.ResolveWorkspace` si solo
     necesita el workspace).

2. **Registralo en el padre**:
   - Si el grupo es `pr`, agregalo en `internal/cmd/pr/pr.go` (`cmd.AddCommand(NewCmdXxx(f))`).
   - Si el grupo es `auth`, `repo`, o `branch` (grupos chicos sin archivo propio),
     agregalo en `internal/cmd/root/root.go` donde se arma ese grupo.
   - Si es un grupo nuevo que no existe todavía, creá el patrón `pr.go` (un
     `NewCmd<Grupo>` que arma el `cobra.Command` padre y le agrega subcomandos)
     y registralo en `root.go`.

3. **Endpoint de API si hace falta**: si el comando necesita un endpoint que
   `internal/api` no tiene, agregalo en `internal/api/<recurso>.go` (creá el
   archivo si el recurso es nuevo) con su struct de respuesta y su método en
   `*Client`. Agregale un test en `internal/api/<recurso>_test.go` usando
   `httptest.Server` y una fixture JSON en `internal/api/testdata/`.

4. **Actualizá el README**: agregá el comando nuevo a la sección "Comandos" de
   `README.md` con un ejemplo de uso.

Al terminar, corré `make build && make lint && make test` y arreglá lo que
falle antes de dar el trabajo por terminado.
