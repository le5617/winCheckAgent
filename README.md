# Agente de Diagnóstico Windows

[![Go Version](https://img.shields.io/badge/Go-1.21%2B-blue)](https://go.dev/)
[![Windows](https://img.shields.io/badge/Windows-10%2F11-blue)](https://www.microsoft.com/windows/)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

Herramienta de diagnóstico avanzada para sistemas Windows que escanea el PC en busca de problemas de hardware y software, lee logs del sistema y ayuda a diagnosticar y prevenir fallos.

## Características

- **Diagnóstico de Hardware**: CPU, memoria, disco, batería, procesos
- **Diagnóstico de Software**: Servicios, drivers, programas de inicio, Windows Update
- **Seguridad**: Integración con Windows Defender, verificación de integridad del sistema
- **Red**: Conectividad, DNS, estado de adaptadores
- **Monitoreo**: Modo tiempo real con alertas
- **Historial**: Registro de diagnósticos para análisis de tendencias

## Instalación

### Opción 1: Descargar binary precompilado

Descarga el archivo `agente_diagnostico.exe` desde la sección [Releases](https://github.com/tu-usuario/agente-diagnostico/releases) y ejecútalo directamente.

### Opción 2: Compilar desde código fuente

```bash
# Clonar el repositorio
git clone https://github.com/tu-usuario/agente-diagnostico.git
cd agente-diagnostico

# Compilar
go build -o agente_diagnostico.exe main.go
```

## Uso

### Línea de comandos

```bash
# Diagnóstico completo
agente_diagnostico.exe

# Diagnóstico rápido
agente_diagnostico.exe --rapido

# Monitoreo en tiempo real (10 segundos)
agente_diagnostico.exe --monitor

# Monitoreo con duración personalizada
agente_diagnostico.exe --monitor 30
```

### Modo interactivo

```bash
agente_diagnostico.exe
```

```
============================================================
  AGENTE DE DIAGNOSTICO WINDOWS - MENU PRINCIPAL
============================================================

1. Diagnostico completo
2. Diagnostico rapido
3. Monitoreo en tiempo real (10s)
4. Ver alertas activas
5. Ver historial
6. Salir
```

## Ejemplo de salida

```
============================================================
  DIAGNOSTICO RAPIDO
============================================================

--- SISTEMA ---
  [+] Sistema: windows amd64
  [+] Hostname: DESKTOP-M3IF5VI
  [+] Plataforma: Microsoft Windows 11 Home
  [+] Tiempo activo: 33m57s
  [+] RAM total: 15.9 GB

--- CPU ---
  [+] Uso promedio: 19.6%
  [+] Modelo: Intel(R) Core(TM) i5-7200U
  [+] Nucleos fisicos: 4

--- MEMORIA ---
  [+] Uso: 48.0% (7.7 GB)

--- MALWARE ---
  [+] Antivirus: ACTIVADO
  [+] Proteccion en tiempo real: ACTIVADA
  [+] No hay amenazas detectadas

============================================================
  RESUMEN: 1 CRITICAL, 1 WARNING
============================================================
```

## Funcionalidades detalladas

| Categoría | Verificaciones |
|-----------|---------------|
| **Hardware** | CPU, memoria, disco, batería, temperatura, SMART |
| **Software** | Windows Update, servicios, drivers, startup |
| **Seguridad** | Windows Defender, integridad (sfc/DISM) |
| **Red** | Conectividad, DNS, adaptadores |
| **Diagnóstico** | BSOD, logs del sistema, errores de disco |

## Recomendaciones post-diagnóstico

```cmd
# Verificar archivos del sistema
sfc /scannow

# Reparar Component Store
DISM /Online /Cleanup-Image /RestoreHealth

# Verificar disco
chkdsk C: /f /r
```

## Archivos generados

- `diagnostico_*.json` - Reporte completo en JSON
- `alertas_activas.json` - Alertas críticas y warnings
- `historial_diagnostico.json` - Historial de diagnósticos (30 días)

## Requisitos

- Windows 10/11
- Go 1.21+ (solo para compilación)
- PowerShell 5.1+

## Por qué Go?

- **Binary standalone**: No requiere runtime ni dependencias
- **Rendimiento**: Monitoreo eficiente en tiempo real
- **Portabilidad**: Fácil compilación cruzada
- **Tamaño**: ~4MB binary final

## Licencia

MIT License - ver [LICENSE](LICENSE) para más detalles.

---

*Nota: Este proyecto requiere permisos de administrador para algunas verificaciones como lectura de logs del sistema y servicios.*
