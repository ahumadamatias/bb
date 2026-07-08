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
go install github.com/matiasahumada/bb/cmd/bb@latest
```

O descargá el binario para tu plataforma desde la [página de releases](https://github.com/matiasahumada/bb/releases).

## Setup

`bb` se autentica con **email + API token de Atlassian** (no con app passwords,
que Atlassian discontinúa a mediados de 2026).

1. Creá un API token en https://id.atlassian.com/manage-profile/security/api-tokens,
   con scopes de lectura/escritura sobre repositorios y pull requests.
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

```sh
bb auth login                        # configura email + API token
bb auth status                       # valida credenciales contra la API

bb repo list                         # lista repos del workspace
bb repo list --workspace acme --limit 50

bb branch list                       # lista ramas del repo actual (infiere workspace/repo del remote git)

bb pr create --title "Mi cambio"     # crea un PR (source = rama actual, dest = rama principal)
bb pr create --title "Fix" --body "Detalle" --reviewer ana,juan --close-source-branch

bb pr list                           # PRs abiertos del repo actual
bb pr list --state MERGED

bb pr view 42                        # detalle de un PR
bb pr view 42 --diff                 # + diff
bb pr view 42 --web                  # abre el PR en el navegador

bb pr comment 42 --body "LGTM"
bb pr approve 42
bb pr approve 42 --remove            # retira la aprobación

bb pr merge 42
bb pr merge 42 --strategy squash --close-source-branch
```

Parado en un clon de un repo de bitbucket.org, `bb` infiere el workspace y el
repo del remote `origin` automáticamente. Fuera de un repo git (o si el remote
no es de Bitbucket), usá `--workspace` y `--repo`.

## Scripting

Cualquier comando de listado o detalle soporta `--output json` (alias `-o json`)
para producir JSON simplificado en lugar de tablas:

```sh
bb pr list --output json | jq -r '.[] | "\(.id)\t\(.title)"'
bb repo list -o json | jq '[.[].slug]'
bb pr view 42 -o json | jq -r '.diff' > 42.diff   # con --diff, el diff va en el campo "diff"
```

Salida por pipe (sin TTY) usa TSV sin encabezados en lugar de tablas alineadas.

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

## Desarrollo

Ver [CLAUDE.md](CLAUDE.md) / [AGENTS.md](AGENTS.md) para la arquitectura, las
reglas del proyecto y cómo agregar un subcomando nuevo.

```sh
make build   # compila bin/bb
make test    # go test ./...
make lint    # go vet + gofmt -l
make install # go install ./cmd/bb
```
