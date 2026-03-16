# Changelog

Todos los cambios notables de este proyecto se documentan en este archivo.

El formato se basa en [Keep a Changelog](https://keepachangelog.com/es-ES/1.0.0/),
y este proyecto se adhiere a [Versionamiento Semantico](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.3.0] - 2026-03-15

### Added

- Metodos Safe de acceso con retorno de error (`StringSafe`, `IntSafe`, `FloatSafe`, `ObjectSafe`, `ArraySafe`, `BoolSafe`) como alternativa segura a los accessors que hacen panic.
- Metodo `Bool()` y `BoolSafe()` para acceso a valores booleanos.
- Fuzz testing con 3 targets (`FuzzUnmarshal`, `FuzzRoundTrip`, `FuzzMarshalString`) usando la API nativa de Go 1.18 `testing.F`.
- Funcion `Marshal` para serializar `Value` a JSON.
- Suite de tests comprehensiva para parser y value.

### Changed

- Refactorizacion del parser y value para mejor legibilidad y rendimiento.
- Los metodos existentes (`String`, `Int`, `Float`, `Object`, `Array`) ahora son wrappers que internamente usan los metodos `*Safe` y hacen panic on error, manteniendo retrocompatibilidad.
- Formato de marshaling de floats cambiado de `'f'` a `'g'` para evitar overflow en round-trip con valores grandes (e.g. `1e20`).
- `go.mod` actualizado de Go 1.12 a Go 1.18 para soporte de fuzzing nativo.

### Fixed

- Bug en marshaling de floats: valores como `1e20` se serializaban como `"100000000000000000000"` (formato `'f'`), causando overflow al re-parsear. Corregido usando formato `'g'`.

## [1.2.0] - 2019-10-10

### Changed

- Actualizado `go.mod` con nuevo modulo.

## [1.1.0] - 2019-10-10

### Added

- Parser JSON con soporte para strings, numeros (int/float), objetos, arrays, booleanos y null.
- Tipo `Value` con accessors tipados (`String()`, `Int()`, `Float()`, `Object()`, `Array()`).

## [1.0.0] - 2019-10-10

### Added

- Commit inicial del proyecto.

[Unreleased]: https://github.com/jarlex/jsonart/compare/v1.3.0...HEAD
[1.3.0]: https://github.com/jarlex/jsonart/compare/v1.2.0...v1.3.0
[1.2.0]: https://github.com/jarlex/jsonart/compare/v1.1.0...v1.2.0
[1.1.0]: https://github.com/jarlex/jsonart/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/jarlex/jsonart/releases/tag/v1.0.0
