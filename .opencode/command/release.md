---
description: Checklist para cortar un release de bb (tag vX.Y.Z + goreleaser)
---

Versión a releasear: `$ARGUMENTS` (ej: "v0.3.0"). Si no viene, preguntá cuál es
antes de seguir.

1. Confirmá que estás en `main` y sin cambios sin commitear (`git status`).
2. Corré `make test` y `make lint`; arreglá cualquier fallo antes de seguir.
3. Confirmá que `$ARGUMENTS` sigue semver y que el tag no existe todavía
   (`git tag -l`).
4. Creá el tag anotado: `git tag -a $ARGUMENTS -m "$ARGUMENTS"`.
5. Pusheá el tag (`git push origin $ARGUMENTS`) — dispara el workflow de
   goreleaser en `.github/workflows/release.yml`.
6. Verificá que el workflow de release terminó bien y que el release en
   GitHub tiene los binarios, checksums y changelog adjuntos.

Pushear tags es visible para todo el mundo con acceso al repo — confirmá con
el usuario antes del paso 5 si hay dudas sobre si corresponde releasear ahora.
