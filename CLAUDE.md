# bb — CLI de Bitbucket Cloud

`bb` es un CLI en Go para Bitbucket Cloud (bitbucket.org, API REST 2.0). No hay
CLI oficial de Atlassian para Bitbucket; este proyecto cubre ese hueco, inspirado
en la arquitectura de `cli/cli` (GitHub CLI) a menor escala.

## Comandos de desarrollo

```sh
make build              # compila ./cmd/bb -> bin/bb
make test               # go test ./...
make lint               # go vet + gofmt -l (falla si hay diffs)
make install            # go install ./cmd/bb

go test ./internal/api/...      # solo el cliente API
go test ./internal/config/...   # solo config
go test ./internal/gitctx/...   # solo gitctx
```

## Arquitectura (5 líneas)

- `cmd/bb/main.go`: arranca `internal/cmd/root`, mapea errores a exit codes (0/1/2/4).
- `internal/cmd/<grupo>/`: un archivo por subcomando de cobra (`auth`, `repo`, `branch`, `pr`). Solo parsean flags.
- `internal/cmd/factory`: construye las dependencias reales (Config, Client, GitContext, IOStreams) una sola vez.
- `internal/api`: cliente HTTP a mano para Bitbucket Cloud 2.0 (auth, paginación, errores). Un archivo por recurso.
- `internal/config`, `internal/gitctx`, `internal/iostreams`, `internal/tableprinter`: config YAML, inferencia de workspace/repo desde git, IO con detección de TTY, tablas.

## Las 4 reglas de oro

1. **Options struct + inyección de dependencias.** Cada subcomando define un struct con todas sus dependencias como funcs/interfaces (`Client func() (*api.Client, error)`, etc.) y una función `xxxRun(opts)` separada del cobra.Command. Nada de variables globales; nunca escribir a `os.Stdout` directo, siempre a `opts.IO.Out` / `opts.IO.ErrOut`.
2. **La lógica vive en `internal/`.** El `cobra.Command` en `internal/cmd/<grupo>/` solo define flags y arma el Options struct; toda la lógica de negocio va en la función `xxxRun`.
3. **Dependencias mínimas.** No agregar librerías sin justificación clara — hoy son cobra, yaml.v3, x/term. Nada de SDKs de terceros para Bitbucket ni frameworks pesados (viper, etc.).
4. **Todo endpoint nuevo en `internal/api` lleva test con `httptest.Server` y fixture en `testdata/`.**

## Agregar un subcomando nuevo

1. Creá `internal/cmd/<grupo>/<accion>.go` con su `XxxOptions` struct y `NewCmdXxx(f *factory.Factory) *cobra.Command`.
2. Registralo en el comando padre (`internal/cmd/<grupo>/<grupo>.go` si existe, o directamente en `internal/cmd/root/root.go` para grupos chicos como `auth`/`repo`/`branch`).
3. Si necesitás un endpoint nuevo, agregalo en `internal/api/<recurso>.go` con su test (regla 4).

Ver también `/new-command` en `.claude/commands/` para el flujo completo.

---
Este archivo debe mantenerse sincronizado con `AGENTS.md`.
