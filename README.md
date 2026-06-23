# Drupal Watcher con Bun

> 🚀 Un watcher inteligente para Drupal que vigila tus módulos y temas custom, ejecutando `drush cr` automáticamente al detectar cambios. Soporte nativo para DDEV, Lando y entornos locales.

[![Licencia MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Bun](https://img.shields.io/badge/Bun-1.3+-black.svg)](https://bun.sh)
[![Composer](https://img.shields.io/badge/Composer-ready-brightgreen.svg)](https://getcomposer.org)

## 📋 Tabla de Contenidos

- [¿Qué hace?](#qué-hace)
- [Características](#características)
- [Requisitos](#requisitos)
- [Instalación](#instalación)
- [Comandos](#comandos)
- [Configuración](#configuración)
- [Ejemplos de uso](#ejemplos-de-uso)
- [Solución de problemas](#solución-de-problemas)
- [Preguntas frecuentes](#preguntas-frecuentes)
- [Contribución](#contribución)
- [Licencia](#licencia)

## ¿Qué hace?

Olvídate de ejecutar manualmente `drush cr` cada vez que modificas un archivo. **Drupal Watcher**:

- **Vigila** en tiempo real los archivos de tus módulos y temas custom
- **Detecta automáticamente** cambios en archivos `.html.twig`, `.inc`, `.yml`, `.module` y `.theme`
- **Ejecuta `drush cr`** de forma inteligente (con debounce para no saturar el sistema)
- **Se adapta** a tu entorno: DDEV, Lando o local
- **Persiste** tus rutas personalizadas en un archivo de configuración

## Características

### Gestión de rutas
- Añade, elimina y lista rutas a vigilar
- Persistencia en `watcher.config.json`
- Validación de existencia de carpetas

### Optimizado para Drupal
- Detecta automáticamente DDEV, Lando o local
- Usa el comando Drush correcto según el entorno
- Debounce inteligente (800ms por defecto)

### Ultra rápido
- Instalación con Bun (10-30x más rápido que npm)
- Arranque en frío instantáneo (~8ms)
- Bajo consumo de memoria

### Desarrollado con Bun 🛠️
- TypeScript/JavaScript moderno
- Sin dependencias externas (solo Bun)
- Ejecutable como binario standalone (opcional)

## Requisitos

- **Bun** (instalado globalmente: `curl -fsSL https://bun.sh/install | bash`)
- **Composer** (gestión de dependencias de PHP)
- **Drupal** con Drush instalado

## Instalación

### Método 1: Desde Packagist (recomendado)

```bash
composer require irving-frias/drupal-watcher
```

### Método 2: Desde repositorio local

1. Clona o descarga el paquete en `packages/drupal-watcher/`
2. Añade el repositorio a tu `composer.json`:

```json
"repositories": [
    {
        "type": "path",
        "url": "packages/drupal-watcher-bun"
    }
]
```

3. Instala:

```bash
composer require irving-frias/drupal-watcher:@dev
```

### Método 3: Desde ZIP

1. Descarga el [ZIP](https://github.com/irving-frias/drupal-watcher/archive/refs/heads/main.zip)
2. Descomprime en `packages/drupal-watcher-bun/`
3. Sigue los pasos del Método 2

## Comandos

Todos los comandos se ejecutan desde la raíz de tu proyecto Drupal usando el binario instalado por Composer.

### Iniciar el watcher

```bash
vendor/bin/drupal-watcher start
```

Verás algo como:

```
🚀 Iniciando Drupal Watcher
📁 Directorio raíz: docroot
🌐 Entorno: ddev
🔧 Drush: ddev drush cr
👀 Vigilando rutas:
  - docroot/modules/custom
  - docroot/themes/custom
✅ Watcher activo. Esperando cambios... (Ctrl+C para salir)
```

### Listar rutas configuradas

```bash
vendor/bin/drupal-watcher list
```

Muestra las rutas actuales, patrones, debounce y comando Drush.

### Añadir una nueva ruta

```bash
vendor/bin/drupal-watcher add docroot/modules/contrib
```

### Eliminar una ruta

```bash
vendor/bin/drupal-watcher remove docroot/modules/contrib
```

### Restablecer rutas por defecto

```bash
vendor/bin/drupal-watcher reset
```

### Comandos como alias de Composer (opcional)

Si prefieres `composer watcher:*`, añade estos scripts a tu `composer.json` raíz:

```json
"scripts": {
    "watcher:start": "vendor/bin/drupal-watcher start",
    "watcher:list": "vendor/bin/drupal-watcher list",
    "watcher:add": "vendor/bin/drupal-watcher add",
    "watcher:remove": "vendor/bin/drupal-watcher remove",
    "watcher:reset": "vendor/bin/drupal-watcher reset"
}
```

Luego podrás ejecutar `composer watcher:start`, etc.

## Configuración

El archivo `watcher.config.json` se crea automáticamente en la raíz de tu proyecto.

### Estructura del archivo

```json
{
  "routes": [
    "docroot/modules/custom",
    "docroot/themes/custom"
  ],
  "patterns": [".html.twig", ".inc", ".yml", ".module", ".theme"],
  "excludePatterns": [],
  "debounce": 800,
  "drushCmd": null,
  "drushCommand": "cr",
  "drushArgs": [],
  "postClearCommands": []
}
```

### Opciones de configuración

| Opción | Descripción | Valor por defecto |
| :--- | :--- | :--- |
| `routes` | Lista de rutas a vigilar | `["docroot/modules/custom", "docroot/themes/custom"]` |
| `patterns` | Extensiones de archivo a vigilar | `[".html.twig", ".inc", ".yml", ".module", ".theme"]` |
| `excludePatterns` | Extensiones a ignorar (ej: `[".css", ".js"]`) | `[]` |
| `debounce` | Tiempo de espera (ms) antes de ejecutar drush | `800` |
| `drushCmd` | Comando Drush personalizado (ej: `"ddev drush"`) | `null` (auto-detecta) |
| `drushCommand` | Subcomando de Drush a ejecutar | `"cr"` |
| `drushArgs` | Argumentos extra para Drush | `[]` |
| `postClearCommands` | Comandos a ejecutar tras limpiar caché | `[]` |

### Notas sobre la configuración

- **Patrones**: Añade o quita extensiones según tus necesidades
- **Exclude patterns**: Ignora archivos como `.css` o `.js` si no quieres que disparen `drush cr`
- **Debounce**: Ajusta según el rendimiento de tu proyecto (proyectos grandes pueden necesitar más tiempo)
- **Drush personalizado**: Si usas un binario Drush en una ubicación específica, defínelo en `drushCmd`
- **drushCommand**: Cambia a `"cc bin"` para limpieza más rápida (solo un bin de caché) en vez de `"cr"` (completo)
- **postClearCommands**: Array de comandos shell a ejecutar tras cada limpieza (ej: `["ddev drush cex", "echo 'listo'"]`)

## Ejemplos de uso

### Ejemplo 1: Watcher básico

```bash
# Instalación
composer require irving-frias/drupal-watcher

# Iniciar watcher
vendor/bin/drupal-watcher start

# Editar un archivo .twig...
📝 mi-plantilla.html.twig
🔄 Limpiando caché...
✅ Caché limpiada correctamente.
```

### Ejemplo 2: Añadir módulos contrib

```bash
# Añadir módulos contrib
vendor/bin/drupal-watcher add docroot/modules/contrib

# Verificar
vendor/bin/drupal-watcher list

# Ahora vigila tanto custom como contrib
```

### Ejemplo 3: Cache clear más rápido

Para limpiar solo un bin en vez de toda la caché, edita `watcher.config.json`:

```json
{
  "drushCommand": "cc bin"
}
```

Esto ejecuta `drush cc bin` en vez de `drush cr`, que es más rápido.

### Ejemplo 4: Comandos post-clear

Ejecuta `drush cex` automáticamente tras cada cambio:

```json
{
  "postClearCommands": ["ddev drush cex"]
}
```

### Ejemplo 5: Compilar a binario standalone

Si no quieres depender de Composer/Bun para el día a día:

```bash
bun build --compile ./vendor/irving-frias/drupal-watcher/bin/drupal-watcher --outfile ./drupal-watcher
./drupal-watcher
```

## Solución de problemas

### ❌ Error: `command not found: bun`

Bun no está instalado globalmente en tu sistema. Verifica:

1. Bun está instalado: `bun --version`
2. Si no lo tienes, instálalo con: `curl -fsSL https://bun.sh/install | bash`

### ❌ Error: `No se encontró Drush`

El watcher busca Drush en:
- `vendor/bin/drush` (proyecto Drupal)
- `bin/drush` (alternativa)
- Comandos de DDEV/Lando

Verifica que:
1. Drush está instalado: `composer require drush/drush`
2. Estás ejecutando desde la raíz del proyecto Drupal

### ❌ Error: `Ninguna de las rutas configuradas existe`

Asegúrate de que:
1. Las carpetas `docroot/modules/custom` y `docroot/themes/custom` existen
2. O añade rutas válidas con `vendor/bin/drupal-watcher add`

### ❌ El watcher no detecta cambios

Verifica:
1. Estás editando archivos con las extensiones correctas (`.html.twig`, `.inc`, `.yml`, `.module`, `.theme`)
2. Los archivos están dentro de las rutas configuradas
3. En proyectos con muchos archivos, el watcher puede tardar en iniciar

### ❌ La caché se limpia demasiado frecuentemente

Aumenta el valor de `debounce` en `watcher.config.json`:

```json
{
  "debounce": 1200  // 1.2 segundos
}
```

## Preguntas frecuentes

### ¿Por qué usar Bun en lugar de Node.js?

Bun es **10-30x más rápido** en instalaciones y **8x más rápido** en arranque en frío. Además, es un "todo-en-uno" (runtime, gestor de paquetes, empaquetador), lo que reduce dependencias y complejidad.

### ¿Puedo usarlo sin Composer?

Sí. Puedes compilar el script a un binario standalone:

```bash
bun build --compile ./bin/drupal-watcher --outfile ./drupal-watcher
./drupal-watcher start
```

### ¿Funciona en Windows?

Sí, Bun es multiplataforma. Puedes instalarlo desde [bun.sh](https://bun.sh).

### ¿Puedo vigilar múltiples proyectos?

No directamente. Cada proyecto tiene su propio watcher y configuración. Ejecuta el comando desde la raíz de cada proyecto.

### ¿Qué archivos vigila?

Por defecto: `.html.twig`, `.inc`, `.yml`, `.module`, `.theme`. Puedes añadir más patrones en `watcher.config.json`.

## Contribución

¡Las contribuciones son bienvenidas!

1. Fork del proyecto
2. Crea una rama para tu feature (`git checkout -b feature/amazing-feature`)
3. Commit tus cambios (`git commit -m 'Add some amazing feature'`)
4. Push a la rama (`git push origin feature/amazing-feature`)
5. Abre un Pull Request

### Reportar issues

Usa el [issue tracker](https://github.com/irving-frias/drupal-watcher/issues) para reportar bugs o sugerir mejoras.

## Licencia

Este proyecto está bajo la Licencia MIT. Ver el archivo [LICENSE](LICENSE) para más detalles.

---

## Agradecimientos

- [Bun](https://bun.sh) - Por su increíble velocidad
- [Drupal](https://www.drupal.org) - Por ser el mejor CMS del mundo

---

**¿Te ha sido útil?** ⭐️ Dale una estrella al repositorio y compártelo con otros desarrolladores de Drupal.

---

*Creado con ❤️ por [Irving Frías](https://github.com/irving-frias)*
