---
description: Corre make test y make lint, y arregla lo que falle hasta que pasen
---

Corré `make test` y `make lint`. Si algo falla, arreglá la causa en el código
fuente (no ajustes tests para forzar que pasen, no uses `gofmt -w` como
sustituto de entender un error de `go vet`) y volvé a correr ambos targets
hasta que terminen sin errores.
