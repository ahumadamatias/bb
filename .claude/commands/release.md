---
description: Checklist para cortar un release de bb (tag vX.Y.Z + goreleaser)
argument-hint: "vX.Y.Z"
---

Versión a releasear: `$ARGUMENTS` (ej: "v0.3.0"). Si no viene, preguntá cuál es
antes de seguir — no la inventes.

Checklist, en orden, sin saltear pasos:

1. Confirmá que estás en `main` y sin cambios sin commitear (`git status`).
2. Corré `make test` y `make lint`; si algo falla, arreglalo antes de seguir
   (no se taggea con tests rotos).
3. Confirmá que `$ARGUMENTS` sigue semver (`vMAJOR.MINOR.PATCH`) y que no existe
   ya ese tag (`git tag -l`).
4. Creá el tag anotado: `git tag -a $ARGUMENTS -m "$ARGUMENTS"`.
5. Pusheá el tag: `git push origin $ARGUMENTS` — esto dispara
   `.github/workflows/release.yml`, que corre goreleaser.
6. Verificá que el workflow de release corrió bien: `gh run list --workflow=release.yml --limit 1`
   y `gh run view <id>` si algo se ve raro.
7. Confirmá que los artifacts (binarios linux/darwin/windows, checksums,
   changelog) quedaron adjuntos al release en GitHub.

Pushear tags y crear releases es una acción visible para todo el mundo con
acceso al repo — confirmá con el usuario antes del paso 5 si hay cualquier
duda sobre si corresponde releasear ahora.
