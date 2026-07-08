---
description: Corre make test y make lint, y arregla lo que falle hasta que pasen
---

Corré `make test` y `make lint` en este repo.

Si alguno falla:

1. Leé el error con cuidado (test que falla, output de `go vet`, o archivos que
   `gofmt -l` marca como no formateados).
2. Arreglá la causa del fallo en el código fuente correspondiente — no ajustes
   el test para que pase si el bug está en el código, y no corras `gofmt -w`
   como sustituto de entender un error real de `go vet`.
3. Volvé a correr `make test` y `make lint`.
4. Repetí hasta que ambos targets terminen sin errores.

Si un test falla de forma consistente y creés que el test mismo está mal
(no el código), explicá por qué antes de tocarlo.
