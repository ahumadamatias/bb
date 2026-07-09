# bb

`bb` es un CLI no oficial para [Bitbucket Cloud](https://bitbucket.org) (API REST 2.0).
Atlassian no tiene un CLI oficial para Bitbucket; `bb` cubre ese hueco para uso
diario en la terminal, scripting y como tool para agentes de código. Su
arquitectura está inspirada en [`cli/cli`](https://github.com/cli/cli) (el CLI
oficial de GitHub), adaptada a una escala menor.

Solo soporta Bitbucket **Cloud** (bitbucket.org). Bitbucket Server / Data Center
no está contemplado.

## Instalación

Con Go instalado:

```sh
go install github.com/ahumadamatias/bb/cmd/bb@latest
```

O descargá el binario para tu plataforma desde la [página de releases](https://github.com/ahumadamatias/bb/releases).

## Setup

`bb` se autentica con **email + API token de Atlassian** (no con app passwords,
que Atlassian discontinúa a mediados de 2026).

1. Creá un **API token with scopes** (no uno clásico sin scopes) en
   https://id.atlassian.com/manage-profile/security/api-tokens, eligiendo
   **Bitbucket** como app. Necesitás al menos estos scopes:

   | Scope | Para qué |
   |---|---|
   | `read:user:bitbucket` | `bb auth login` / `bb auth status` (valida contra `GET /user`) |
   | `read:workspace:bitbucket` | listar workspaces al hacer login, resolver `--reviewer` por nombre |
   | `read:repository:bitbucket` | `bb repo list`, `bb branch list`, `bb pr create` (lee la rama principal) |
   | `read:pullrequest:bitbucket`, `write:pullrequest:bitbucket` | `bb pr list/view/create/comment/approve/merge` |

   Si te falta alguno, la API devuelve un 403 con el mensaje
   `"Your credentials lack one or more required privilege scopes"` indicando
   exactamente qué scope falta (`error.detail.required` en el JSON de la
   respuesta) — es la forma más confiable de diagnosticarlo si algo falla.

2. Corré:

   ```sh
   bb auth login
   ```

   Sin flags, pregunta el email y el token de forma interactiva (el token no se
   muestra en pantalla) y ofrece elegir un workspace por defecto. También podés
   pasar las credenciales directo:

   ```sh
   bb auth login --email vos@example.com --token TU_API_TOKEN
   ```

3. Verificá que quedó bien configurado:

   ```sh
   bb auth status
   ```

Las credenciales se guardan en `~/.config/bb/config.yaml` (o `$XDG_CONFIG_HOME/bb/config.yaml`)
con permisos `0600`.

## Comandos

Parado en un clon de un repo de bitbucket.org, `bb` infiere `workspace`, `repo`
y rama actual desde el remote `origin`. Soporta remotes SSH y HTTPS, por ejemplo
`git@bitbucket.org:workspace/repo.git` y
`https://bitbucket.org/workspace/repo.git`.

Fuera de un repo git, o si el remote no es de Bitbucket, pasá `--workspace` y
`--repo` donde corresponda:

```sh
bb pr list --workspace acme --repo api
bb branch list -w acme --repo api
```

Flags globales disponibles en todos los comandos:

| Flag | Alias | Variable | Para qué |
|---|---:|---|---|
| `--email` | | `BB_EMAIL` | Email de Atlassian usado para autenticar |
| `--token` | | `BB_TOKEN` | API token de Atlassian |
| `--workspace` | `-w` | `BB_WORKSPACE` | Workspace de Bitbucket; si no se pasa, se infiere o usa el default |
| `--repo` | | | Repo slug de Bitbucket; si no se pasa, se infiere del remote git |
| `--output json` | `-o json` | | Salida JSON para comandos de listado/detalle |

Precedencia de configuración: flags > variables de entorno >
`~/.config/bb/config.yaml`.

### Resumen rápido

```sh
bb auth login                        # configura email + API token
bb auth status                       # valida credenciales contra la API

bb repo list                         # lista repos del workspace
bb repo list --workspace acme --limit 50

bb branch list                       # lista ramas del repo actual (infiere workspace/repo del remote git)

bb pr create --title "Mi cambio"     # crea un PR (source = rama actual, dest = rama principal)
bb pr create --title "Fix" --body "Detalle" --reviewer ana,juan --close-source-branch

bb pr list                           # PRs abiertos del repo actual (--state acepta un solo valor: OPEN, MERGED o DECLINED)
bb pr list --state MERGED
bb pr list --state DECLINED

bb pr view 42                        # detalle de un PR
bb pr view 42 --diff                 # + diff
bb pr view 42 --web                  # abre el PR en el navegador

bb pr comment 42 --body "LGTM"                                          # comentario general
bb pr comment 42 --body "esto está mal" --path internal/api/client.go --line 42   # comentario inline
bb pr comment 42 --body "hay que arreglar esto" --path a/b.go --line 10 --task    # + marcado como task bloqueante

bb pr approve 42
bb pr approve 42 --remove            # retira la aprobación

bb pr merge 42
bb pr merge 42 --strategy squash --close-source-branch
```

### `bb auth login`

Configura credenciales en `~/.config/bb/config.yaml` o en
`$XDG_CONFIG_HOME/bb/config.yaml` si `XDG_CONFIG_HOME` está seteado. El archivo
se escribe con permisos `0600`.

```sh
bb auth login
bb auth login --email vos@example.com --token TU_API_TOKEN
```

Sin flags, pide email y token de forma interactiva. Si hay una terminal
interactiva, después de validar credenciales intenta listar workspaces y permite
elegir uno como workspace por defecto. En scripts o CI, pasá `--email` y
`--token`, o usá `BB_EMAIL` y `BB_TOKEN`.

### `bb auth status`

Valida las credenciales resueltas contra `GET /user` y muestra el usuario y el
workspace por defecto, si existe.

```sh
bb auth status
```

### `bb repo list`

Lista repositorios del workspace resuelto.

```sh
bb repo list
bb repo list --workspace acme
bb repo list --workspace acme --limit 50
bb repo list --limit 0              # todos los repositorios
bb repo list -o json | jq '.[].slug'
```

Flags:

| Flag | Default | Para qué |
|---|---:|---|
| `--limit` | `30` | Máximo de repositorios a traer; `0` trae todos |

Salida texto: tabla con `NAME`, `DESCRIPTION`, `UPDATED`, `PRIVATE`.

### `bb branch list`

Lista ramas del repo resuelto.

```sh
bb branch list
bb branch list --workspace acme --repo api
bb branch list --limit 100
bb branch list --limit 0            # todas las ramas
bb branch list -o json | jq '.[].name'
```

Flags:

| Flag | Default | Para qué |
|---|---:|---|
| `--limit` | `30` | Máximo de ramas a traer; `0` trae todas |

Salida texto: tabla con `NAME`, `LAST COMMIT`, `AUTHOR`, `DATE`.

### `bb pr create`

Crea un pull request en el repo resuelto.

```sh
bb pr create --title "Mi cambio"
bb pr create --title "Fix" --body "Detalle del fix"
bb pr create --title "Fix" --source feature/fix --dest main
bb pr create --title "Fix" --reviewer ana --reviewer juan
bb pr create --title "Fix" --reviewer ana,juan --close-source-branch --web
```

Defaults:

| Campo | Cómo se resuelve |
|---|---|
| `--source` | Rama git actual |
| `--dest` | Rama principal configurada en Bitbucket para el repo |

Flags:

| Flag | Para qué |
|---|---|
| `--title` | Título del PR. Si falta y hay TTY, lo pide interactivo |
| `--body` | Descripción del PR |
| `--source` | Rama origen |
| `--dest` | Rama destino |
| `--reviewer` | Reviewer por nickname, display name, username o UUID `{...}`; repetible o separado por comas |
| `--close-source-branch` | Pide cerrar la rama origen al mergear |
| `--web` | Abre el PR creado en el navegador |

En contexto no interactivo, `--title` es obligatorio.

### `bb pr list`

Lista pull requests del repo resuelto.

```sh
bb pr list
bb pr list --state OPEN
bb pr list --state MERGED --limit 100
bb pr list --state DECLINED --limit 0
bb pr list -o json | jq '.[].id'
```

Flags:

| Flag | Default | Para qué |
|---|---:|---|
| `--state` | `OPEN` | Estado a listar: `OPEN`, `MERGED` o `DECLINED` |
| `--limit` | `30` | Máximo de PRs a traer; `0` trae todos |

`bb pr list` no tiene `--state ALL`; para traer varios estados combinados,
ejecutalo una vez por estado.

Salida texto: tabla con `ID`, `TITLE`, `SOURCE → DEST`, `AUTHOR`, `STATE`.

### `bb pr view <id>`

Muestra el detalle de un pull request.

```sh
bb pr view 42
bb pr view 42 --diff
bb pr view 42 --web
bb pr view 42 --diff -o json | jq -r '.diff'
```

Flags:

| Flag | Para qué |
|---|---|
| `--diff` | Incluye el diff del PR |
| `--web` | Abre el PR en el navegador y no imprime detalle |

Con `-o json`, `--diff` agrega el campo `diff` al JSON.

### `bb pr comment <id>`

Agrega un comentario general o inline a un pull request.

```sh
bb pr comment 42 --body "LGTM"
bb pr comment 42 --body "revisar esto" --path internal/api/client.go --line 42
bb pr comment 42 --body "bloqueante" --path internal/api/client.go --line 42 --task
```

Flags:

| Flag | Para qué |
|---|---|
| `--body` | Texto del comentario; obligatorio |
| `--path` | Archivo del diff donde crear el comentario inline |
| `--line` | Línea en la versión nueva del archivo; requiere `--path` |
| `--task` | Crea además una task bloqueante asociada al comentario |

`--task` marca el comentario como una **task** de Bitbucket (aparece en la lista
de tasks del PR, con estado `UNRESOLVED`) — útil como filtro para detectar
comentarios bloqueantes que hay que resolver antes de mergear, a diferencia de
un comentario normal que es solo informativo.

### `bb pr approve <id>`

Aprueba un PR o retira tu aprobación.

```sh
bb pr approve 42
bb pr approve 42 --remove
```

Flags:

| Flag | Para qué |
|---|---|
| `--remove` | Retira tu aprobación en lugar de aprobar |

### `bb pr merge <id>`

Mergea un pull request abierto.

```sh
bb pr merge 42
bb pr merge 42 --strategy squash
bb pr merge 42 --strategy fast-forward --close-source-branch
```

Flags:

| Flag | Default | Para qué |
|---|---:|---|
| `--strategy` | `merge-commit` | Estrategia: `merge-commit`, `squash` o `fast-forward` |
| `--close-source-branch` | `false` | Cierra la rama origen al mergear |

El comando verifica antes que el PR esté en estado `OPEN`.

## Scripting

Los comandos de listado o detalle soportan `--output json` (alias `-o json`) para
producir JSON en lugar de tablas:

```sh
bb pr list --output json | jq -r '.[] | "\(.id)\t\(.title)"'
bb repo list -o json | jq '[.[].slug]'
bb pr view 42 -o json | jq -r '.diff' > 42.diff   # con --diff, el diff va en el campo "diff"

# bb pr list no tiene --state ALL: para traer los tres estados combinados
for s in OPEN MERGED DECLINED; do bb pr list --state $s --output json; done | jq -s 'add'
```

Salida por pipe (sin TTY) usa TSV sin encabezados en lugar de tablas alineadas
cuando no se pidió JSON. Esto facilita combinar `bb` con `cut`, `awk`, `fzf` o
scripts simples.

## Variables de entorno

| Variable        | Equivale a          |
|-----------------|----------------------|
| `BB_EMAIL`      | `--email`            |
| `BB_TOKEN`      | `--token`             |
| `BB_WORKSPACE`  | `--workspace`         |

Precedencia: flags > variables de entorno > `~/.config/bb/config.yaml`.

## Exit codes

| Code | Significado                          |
|------|----------------------------------------|
| 0    | Éxito                                  |
| 1    | Error general                          |
| 2    | Error de uso (flags/argumentos inválidos) |
| 4    | Error de autenticación / sin credenciales |

## Troubleshooting

### "HTTP 403: Your credentials lack one or more required privilege scopes"

Tu token no tiene el scope que ese comando específico necesita. Bitbucket te
dice exactamente cuál en el propio error — pegá el JSON crudo de la respuesta
(no hace falta `bb` para esto, un `curl` alcanza) y mirá el campo
`error.detail.required`:

```sh
curl -s -u "TU_EMAIL:TU_TOKEN" https://api.bitbucket.org/2.0/user | python3 -m json.tool
```

Si te falta un scope, no se puede agregar a un token ya creado — hay que crear
uno nuevo en id.atlassian.com con todos los de la tabla de la sección
[Setup](#setup) tildados (ojo con tildar **tanto Read como Write** en Pull
Requests; es común tildar solo uno de los dos).

### "HTTP 410: CHANGE-2770 - Functionality has been deprecated"

Si estás modificando `bb` y agregás una llamada a `GET /2.0/workspaces` o
`GET /2.0/user/permissions/workspaces`, vas a pegar con esto: Atlassian
eliminó esos endpoints (cross-workspace listing sin scope de usuario). El
reemplazo vigente es `GET /2.0/user/workspaces` (usado internamente por
`ListWorkspaces` en `internal/api/workspaces.go`), cuya respuesta además viene
anidada (`values[].workspace.{uuid,slug}`, sin campo `name`) — no la forma
plana del endpoint viejo.

### Nunca compartas tu API token en texto plano

Si necesitás depurar una llamada a la API con ayuda de un agente de código (o
de otra persona), no pegues el token en el chat — quien lo tenga puede usarlo
hasta que expire o lo revoques. Usá un script que lo pida de forma interactiva
(`read -rsp`) y guarde solo la respuesta (sin el token) en un archivo aparte
para compartir. Si ya expusiste un token por error, revocalo en
id.atlassian.com y generá uno nuevo.

## Desarrollo

Ver [CLAUDE.md](CLAUDE.md) / [AGENTS.md](AGENTS.md) para la arquitectura, las
reglas del proyecto y cómo agregar un subcomando nuevo.

```sh
make build   # compila bin/bb
make test    # go test ./...
make lint    # go vet + gofmt -l
make install # go install ./cmd/bb
```
