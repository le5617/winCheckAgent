# Ciberox Endpoint - Agente de Diagnóstico Windows

[![Go Version](https://img.shields.io/badge/Go-1.21%2B-blue)](https://go.dev/)
[![Windows](https://img.shields.io/badge/Windows-10%2F11-blue)](https://www.microsoft.com/windows/)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

Herramienta de auditoría y diagnóstico para sistemas Windows que escanea el equipo en busca de problemas de hardware, software, y seguridad. Genera reportes JSON estructurados para integración con sistemas de gestión.

## Características

- **Auditoría de Hardware**: CPU, memoria, disco (S.M.A.R.T.), batería (ciclos de vida), BIOS
- **Auditoría de Software**: Inventario de apps, autorun, extensiones de navegador, vulnerabilidades conocidas (CVE)
- **Auditoría de Seguridad**: Puertos, parches, usuarios, USB, firewall, políticas de contraseña
- **Inventario Patrimonial**: Número de serie, modelo, fabricante para gestión de activos
- **Reportes JSON**: Estructura anidada para integración con SIEM/SOAR

## Instalación

### Opción 1: Descargar binario precompilado

Descarga el archivo `ciberox-endpoint.exe` desde la sección [Releases](https://github.com/le5617/winCheckAgent/releases) y ejecútalo directamente.

### Opción 2: Compilar desde código fuente

```bash
# Clonar el repositorio
git clone https://github.com/le5617/winCheckAgent.git
cd winCheckAgent

# Compilar
go build -o ciberox-endpoint.exe main.go
```

## Uso

Ejecuta el binario para iniciar un escaneo completo automáticamente:

```bash
./ciberox-endpoint.exe
```

El escaneo genera un archivo JSON con toda la información recopilada.

## Permisos Requeridos

Este agente requiere ciertos permisos del sistema para funcionar correctamente:

### Permisos de Administrador (Recomendado)

Algunas verificaciones requieren privilegios de administrador:

| Función | Permiso Requerido | Razón |
|---------|-------------------|-------|
| S.M.A.R.T. del disco | Administrador | Leer atributos SMART de almacenamiento |
| Logs de eventos | Administrador | Leer logs de Windows Event Viewer |
| Servicios del sistema | Administrador | Ver estado de todos los servicios |
| Políticas de seguridad | Administrador | Ver políticas de contraseña y usuarios |

### Permisos de Usuario Normal

Las siguientes funciones funcionan con permisos de usuario:

| Función | Permiso Requerido | Razón |
|---------|-------------------|-------|
| Información del equipo (WMI) | Usuario | Win32_ComputerSystem es legible |
| Hardware básico | Usuario | CPU, memoria, disco básico |
| Aplicaciones instaladas | Usuario | Registro de usuario actual |
| Extensiones de navegador | Usuario | Archivos locales del perfil |

### WMI (Windows Management Instrumentation)

**Se requiere acceso a WMI** para leer información del sistema:

- `Win32_ComputerSystem` - Nombre del equipo, dominio, modelo
- `Win32_OperatingSystem` - Versión del SO
- `Win32_BIOS` - Número de serie, versión BIOS
- `Win32_Processor` - Info del CPU
- `Win32_PhysicalMemory` - RAM instalada
- `Win32_Battery` - Estado de batería
- `Win32_DiskDrive` - Discos físicos

**Nota**: WMI es accesible por defecto para usuarios autenticados en Windows.

### PowerShell

Se utiliza PowerShell para funciones avanzadas:

- **Firewall**: Se requiere acceso a `Get-NetFirewallProfile` y `Get-NetFirewallRule`
- **HotFixes**: `Get-HotFix` para parches instalados
- **USB**: Lectura del registro `HKLM:\SYSTEM\CurrentControlSet\Enum\USBSTOR\`
- **Autorun**: Claves del registro `Run`, `RunOnce`

## Estructura del JSON

El reporte genera un JSON estructurado:

```json
{
  "fecha": "2024-01-15 10:30:00",
  "equipo": {
    "nombre_equipo": "DESKTOP-XXXX",
    "dominio": "WORKGROUP",
    "modelo": "Dell Latitude 5420",
    "numero_serie": "ABC123456"
  },
  "hardware": {
    "discos": [...],
    "bateria": {...},
    "bios": {...},
    "procesador": {...},
    "ram": {...}
  },
  "sistema": {
    "cpu_porcentaje": 25.5,
    "memoria_porcentaje": 60.0,
    "disco_porcentaje": 75.0
  },
  "seguridad": {
    "puertos_escucha": [...],
    "parches_seguridad": [...],
    "usuarios_locales": [...],
    "firewall": {...},
    "software": {
      "aplicaciones_instaladas": [...],
      "autorun": [...],
      "extensiones_navegador": {...},
      "vulnerabilidades": [...]
    }
  },
  "alertas": [...]
}
```

## Funcionalidades Detalladas

### Auditoría de Hardware (Ciclo de Vida)

| Función | Descripción | Permiso | Método |
|---------|-------------|---------|--------|
| `getWMIComputerInfo` | Nombre, dominio, rol del equipo | Usuario | WMI |
| `getWMIProcessorInfo` | Modelo, núcleos, hilos, frecuencia | Usuario | WMI |
| `getWMIPhysicalMemory` | RAM total, slots utilizados, velocidad | Usuario | WMI |
| `getWMIBatteryInfo` | Carga, salud %, capacidad diseño vs actual | Usuario | WMI |
| `getWMIDiskDrive` | Modelo, número de serie del disco | Usuario | WMI |
| `getDiskHealth` | Lectura S.M.A.R.T. (temperatura, errores) | Admin | PowerShell |
| `getBIOSInfo` | Versión BIOS, fecha, fabricante | Usuario | WMI |

### Auditoría de Software (Compliance)

| Función | Descripción | Permiso | Método |
|---------|-------------|---------|--------|
| `getInstalledSoftware` | Lista todas las apps instaladas del registro | Usuario | PowerShell |
| `getAutorunEntries` | Claves Run/RunOnce del registro | Usuario | PowerShell |
| `getBrowserExtensions` | Extensiones de Chrome y Edge | Usuario | PowerShell |
| `checkVulnerabilities` | Detecta versiones vulnerables (Adobe, Java, etc.) | - | Código interno |

### Auditoría de Seguridad (Ciberseguridad)

| Función | Descripción | Permiso | Método |
|---------|-------------|---------|--------|
| `getListeningPorts` | Puertos TCP en modo LISTEN con proceso | Usuario | Go net |
| `getSecurityPatches` | HotFixes instalados, días sin parchar | Usuario | PowerShell |
| `getPasswordPolicy` | Verifica si es admin, lista administradores | Admin | PowerShell |
| `getUserList` | Usuarios locales y su estado | Usuario | PowerShell |
| `getUSBDevices` | Dispositivos USB conectados e historial | Usuario | PowerShell |
| `getFirewallStatus` | Estado del firewall y reglas peligrosas | Admin | PowerShell |

### Información del Sistema

| Función | Descripción | Permiso | Método |
|---------|-------------|---------|--------|
| `getSystemInfo` | Plataforma, kernel, uptime, RAM | Usuario | gopsutil |
| `checkCPU` | Uso de CPU por core | Usuario | gopsutil |
| `checkMemory` | Uso de memoria y swap | Usuario | gopsutil |
| `checkDisk` | Espacio en disco por unidad | Usuario | gopsutil |
| `checkNetwork` | Interfaces de red y conexiones | Usuario | gopsutil |
| `checkBattery` | Estado de carga de batería | Usuario | gopsutil |
| `checkProcesses` | Top procesos por CPU/memoria | Usuario | gopsutil |
| `checkServices` | Servicios de Windows y su estado | Admin | PowerShell |

### Diagnóstico Avanzado

| Función | Descripción | Permiso | Método |
|---------|-------------|---------|--------|
| `checkBSOD` | Analiza pantallas azules en logs | Admin | PowerShell |
| `checkSMART` | Verifica estado SMART del disco | Admin | PowerShell |
| `checkMalware` | Estado de Windows Defender | Usuario | PowerShell |
| `checkDrivers` | Verifica drivers con problemas | Admin | PowerShell |
| `checkSystemIntegrity` | Ejecuta sfc /scannow | Admin | PowerShell |
| `checkWindowsUpdate` | Estado de Windows Update | Usuario | PowerShell |
| `checkLogs` | Lee errores/advertencias del sistema | Admin | PowerShell |

## Archivos Generados

- `diagnostico_YYYYMMDD_HHMMSS.json` - Reporte completo en JSON

## Requisitos

- Windows 10/11
- Go 1.21+ (solo para compilación)
- PowerShell 5.1+
- **Recomendado**: Ejecutar como administrador para escaneo completo

## Por qué Go?

- **Binary standalone**: No requiere runtime ni dependencias
- **Rendimiento**: Escaneo eficiente y rápido
- **WMI nativo**: Librería `github.com/StackExchange/wmi` para acceso directo
- **gopsutil**: Acceso multiplataforma a métricas del sistema

## Licencia

MIT License - ver [LICENSE](LICENSE) para más detalles.

---

*Este proyecto requiere permisos de administrador para verificaciones completas de seguridad, hardware avanzado y logs del sistema.*
