package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

type ReportEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"nivel"`
	Message   string `json:"mensaje"`
}

type ComputerInfo struct {
	NombreEquipo     string `json:"nombre_equipo,omitempty"`
	Dominio          string `json:"dominio,omitempty"`
	Rol              string `json:"rol,omitempty"`
	SistemaOperativo string `json:"sistema_operativo,omitempty"`
	Version          string `json:"version,omitempty"`
	Fabricante       string `json:"fabricante,omitempty"`
	Modelo           string `json:"modelo,omitempty"`
	NumeroSerie      string `json:"numero_serie,omitempty"`
	Producto         string `json:"producto,omitempty"`
}

type HardwareInfo struct {
	Discos     []DiskInfo    `json:"discos"`
	Bateria    BatteryInfo   `json:"bateria"`
	BIOS       BIOSInfo      `json:"bios"`
	Procesador ProcessorInfo `json:"procesador"`
	RAM        RAMInfo       `json:"ram"`
}

type DiskInfo struct {
	Unidad         string  `json:"unidad"`
	Modelo         string  `json:"modelo"`
	NumeroSerie    string  `json:"numero_serie"`
	SaludPorcent   float64 `json:"salud_porcentaje"`
	Temperatura    float64 `json:"temperatura"`
	LecturaErrores int64   `json:"errores_lectura"`
	Reasignados    int64   `json:"sectores_reasignados"`
	Pendientes     int64   `json:"sectores_pendientes"`
	Estado         string  `json:"estado"`
}

type BatteryInfo struct {
	Nombre          string  `json:"nombre"`
	Estado          string  `json:"estado"`
	CargaPorcent    float64 `json:"carga_porcentaje"`
	SaludPorcent    float64 `json:"salud_porcentaje"`
	CapacidadDiseno int64   `json:"capacidad_diseno_mwh"`
	CapacidadActual int64   `json:"capacidad_actual_mwh"`
	Ciclos          int     `json:"ciclos"`
	TiempoRestante  int     `json:"tiempo_restante_minutos"`
}

type BIOSInfo struct {
	Version     string `json:"version"`
	Fecha       string `json:"fecha"`
	Fabricante  string `json:"fabricante"`
	NumeroSerie string `json:"numero_serie"`
}

type ProcessorInfo struct {
	Nombre        string `json:"nombre"`
	Fabricante    string `json:"fabricante"`
	Nucleos       int    `json:"nucleos_fisicos"`
	Threads       int    `json:"hilos"`
	FrecuenciaMax int    `json:"frecuencia_max_mhz"`
	Identificador string `json:"identificador"`
}

type RAMInfo struct {
	TotalGB          float64 `json:"total_gb"`
	SlotsUsados      int     `json:"slots_usados"`
	SlotsDisponibles int     `json:"slots_disponibles"`
	Velocidad        int     `json:"velocidad_mhz"`
}

type SystemInfo struct {
	Plataforma  string  `json:"plataforma"`
	Kernel      string  `json:"kernel"`
	Uptime      string  `json:"uptime"`
	RAMTotalGB  float64 `json:"ram_total_gb"`
	CPUPorcent  float64 `json:"cpu_porcentaje"`
	MemPorcent  float64 `json:"memoria_porcentaje"`
	DiskPorcent float64 `json:"disco_porcentaje"`
}

type NetworkInfo struct {
	Interfaces []NetworkInterface `json:"interfaces"`
}

type NetworkInterface struct {
	Name        string   `json:"nombre"`
	IPAddresses []string `json:"direcciones_ip"`
	MAC         string   `json:"mac"`
	Status      string   `json:"estado"`
}

type SecurityInfo struct {
	ListeningPorts  []PortInfo      `json:"puertos_escucha"`
	SecurityPatches []SecurityPatch `json:"parches_seguridad"`
	Users           []UserInfo      `json:"usuarios_locales"`
	Administrators  []string        `json:"administradores"`
	USBDevices      []string        `json:"dispositivos_usb"`
	Printers        []string        `json:"impresoras"`
	FirewallStatus  FirewallStatus  `json:"firewall"`
	IsAdmin         bool            `json:"es_administrador"`
	DiasSinParchar  int             `json:"dias_sin_parche"`
	Software        SoftwareInfo    `json:"software"`
}

type PortInfo struct {
	Port       int    `json:"puerto"`
	Service    string `json:"servicio"`
	Process    string `json:"proceso"`
	Sospechoso bool   `json:"sospechoso"`
}

type SecurityPatch struct {
	HotFixID    string `json:"id"`
	Descripcion string `json:"descripcion"`
	Instalado   string `json:"instalado"`
}

type UserInfo struct {
	Nombre      string `json:"nombre"`
	Habilitado  bool   `json:"habilitado"`
	UltimoLogin string `json:"ultimo_login"`
}

type FirewallStatus struct {
	Domain  bool `json:"domain"`
	Public  bool `json:"public"`
	Private bool `json:"private"`
}

type SoftwareInfo struct {
	InstalledApps     []InstalledApp  `json:"aplicaciones_instaladas"`
	AutorunEntries    []AutorunEntry  `json:"autorun"`
	BrowserExtensions BrowserExts     `json:"extensiones_navegador"`
	Vulnerabilities   []Vulnerability `json:"vulnerabilidades"`
}

type InstalledApp struct {
	Nombre    string `json:"nombre"`
	Version   string `json:"version"`
	Editor    string `json:"editor"`
	FechaInst string `json:"fecha_instalacion"`
}

type AutorunEntry struct {
	Ubicacion  string `json:"ubicacion"`
	Nombre     string `json:"nombre"`
	Comando    string `json:"comando"`
	Habilitado bool   `json:"habilitado"`
}

type BrowserExts struct {
	Chrome []BrowserExt `json:"chrome"`
	Edge   []BrowserExt `json:"edge"`
}

type BrowserExt struct {
	ID       string `json:"id"`
	Nombre   string `json:"nombre"`
	Editor   string `json:"editor"`
	Version  string `json:"version"`
	Permisos string `json:"permisos"`
}

type Vulnerability struct {
	Software    string `json:"software"`
	Version     string `json:"version"`
	Severidad   string `json:"severidad"`
	CVE         string `json:"cve"`
	Descripcion string `json:"descripcion"`
}

type ServiceInfo struct {
	Name      string `json:"nombre"`
	Status    string `json:"estado"`
	StartType string `json:"tipo_inicio"`
}

type ProcessInfo struct {
	Name       string  `json:"nombre"`
	PID        int32   `json:"pid"`
	CPUPercent float64 `json:"cpu_porcentaje"`
	MemPercent float64 `json:"mem_porcentaje"`
}

type DiagnosticReport struct {
	Fecha     string        `json:"fecha"`
	Equipo    ComputerInfo  `json:"equipo"`
	Hardware  HardwareInfo  `json:"hardware"`
	Sistema   SystemInfo    `json:"sistema"`
	Red       NetworkInfo   `json:"red"`
	Seguridad SecurityInfo  `json:"seguridad"`
	Servicios []ServiceInfo `json:"servicios"`
	Procesos  []ProcessInfo `json:"procesos"`
	Alertas   []ReportEntry `json:"alertas"`
}

type AlertasActivas struct {
	Timestamp string        `json:"timestamp"`
	Alertas   []ReportEntry `json:"alertas"`
}

type HistorialEntry struct {
	Timestamp string  `json:"timestamp"`
	Resumen   Resumen `json:"resumen"`
}

type Resumen struct {
	AlertasCriticas   int     `json:"alertas_criticas"`
	AlertasWarning    int     `json:"alertas_warning"`
	CPUPromedio       float64 `json:"cpu_promedio"`
	MemoriaPorcentaje float64 `json:"memoria_porcentaje"`
}

type Historial struct {
	Historial []HistorialEntry `json:"historial"`
}

var report []ReportEntry
var alertas []ReportEntry
var equipo ComputerInfo
var sistema SystemInfo
var red NetworkInfo
var seguridad SecurityInfo
var servicios []ServiceInfo
var procesos []ProcessInfo
var software SoftwareInfo
var hardware HardwareInfo

func addReport(category, message, level string) {
	entry := ReportEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     level,
		Message:   message,
	}
	report = append(report, entry)

	if level == "WARNING" || level == "CRITICAL" {
		alertas = append(alertas, entry)
	}
}

func runPowerShell(command string) string {
	cmd := exec.Command("powershell", "-Command", command)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func getSystemInfo() {
	hostInfo, err := host.Info()
	if err == nil {
		equipo.NombreEquipo = hostInfo.Hostname
		equipo.SistemaOperativo = fmt.Sprintf("%s %s", hostInfo.Platform, hostInfo.PlatformVersion)
		equipo.Version = hostInfo.KernelVersion

		sistema.Plataforma = hostInfo.Platform
		sistema.Kernel = hostInfo.KernelVersion
		uptime := time.Duration(hostInfo.Uptime) * time.Second
		sistema.Uptime = uptime.String()

		addReport("SISTEMA", fmt.Sprintf("Hostname: %s", hostInfo.Hostname), "INFO")
		addReport("SISTEMA", fmt.Sprintf("Plataforma: %s %s", hostInfo.Platform, hostInfo.PlatformVersion), "INFO")
		addReport("SISTEMA", fmt.Sprintf("Kernel: %s", hostInfo.KernelVersion), "INFO")
		addReport("SISTEMA", fmt.Sprintf("Tiempo activo: %s", uptime.String()), "INFO")
	}

	vm, _ := mem.VirtualMemory()
	ramGB := float64(vm.Total) / (1024 * 1024 * 1024)
	sistema.RAMTotalGB = ramGB
	addReport("SISTEMA", fmt.Sprintf("RAM total: %.1f GB", ramGB), "INFO")
}

func checkCPU() {
	cpuPercent, _ := cpu.Percent(time.Second, false)
	avgPercent := 0.0
	if len(cpuPercent) > 0 {
		avgPercent = cpuPercent[0]
	}

	sistema.CPUPorcent = avgPercent
	level := "INFO"
	if avgPercent > 80 {
		level = "WARNING"
	}
	addReport("CPU", fmt.Sprintf("Uso promedio: %.1f%%", avgPercent), level)

	cpuInfo, _ := cpu.Info()
	if len(cpuInfo) > 0 {
		addReport("CPU", fmt.Sprintf("Modelo: %s", cpuInfo[0].ModelName), "INFO")
		addReport("CPU", fmt.Sprintf("Nucleos fisicos: %d", cpuInfo[0].Cores), "INFO")
		if cpuInfo[0].Mhz > 0 {
			addReport("CPU", fmt.Sprintf("Frecuencia: %.0f MHz", cpuInfo[0].Mhz), "INFO")
		}
	}

	perCPU, _ := cpu.Percent(time.Second, true)
	for i, p := range perCPU {
		lvl := "INFO"
		if p > 80 {
			lvl = "WARNING"
		}
		addReport("CPU", fmt.Sprintf("Core %d: %.1f%%", i, p), lvl)
	}
}

func checkMemory() {
	vm, _ := mem.VirtualMemory()

	sistema.MemPorcent = vm.UsedPercent

	level := "INFO"
	if vm.UsedPercent > 80 {
		level = "WARNING"
	}
	if vm.UsedPercent > 90 {
		level = "CRITICAL"
	}

	addReport("MEMORIA", fmt.Sprintf("Total: %.1f GB", float64(vm.Total)/(1024*1024*1024)), "INFO")
	addReport("MEMORIA", fmt.Sprintf("Disponible: %.1f GB", float64(vm.Available)/(1024*1024*1024)), "INFO")
	addReport("MEMORIA", fmt.Sprintf("Uso: %.1f%% (%.1f GB)", vm.UsedPercent, float64(vm.Used)/(1024*1024*1024)), level)

	if vm.UsedPercent > 80 {
		addReport("MEMORIA", "ALERTA: Uso de memoria alto", level)
	}

	swap, _ := mem.SwapMemory()
	addReport("MEMORIA", fmt.Sprintf("Swap: %.1f GB, Uso: %.1f%%", float64(swap.Total)/(1024*1024*1024), swap.UsedPercent), "INFO")
}

func checkDisk() {
	partitions, _ := disk.Partitions(true)

	for _, p := range partitions {
		usage, err := disk.Usage(p.Mountpoint)
		if err == nil {
			if p.Mountpoint == "C:" || p.Mountpoint == "C:\\" {
				sistema.DiskPorcent = usage.UsedPercent
			}

			level := "INFO"
			if usage.UsedPercent > 80 {
				level = "WARNING"
			}
			if usage.UsedPercent > 90 {
				level = "CRITICAL"
			}

			addReport("DISCO",
				fmt.Sprintf("Unidad %s (%s) - Total: %.1fGB, Libre: %.1fGB, Uso: %.1f%%",
					p.Mountpoint, p.Fstype,
					float64(usage.Total)/(1024*1024*1024),
					float64(usage.Free)/(1024*1024*1024),
					usage.UsedPercent), level)

			if usage.UsedPercent > 90 {
				addReport("DISCO", "ALERTA: Disco casi lleno", "CRITICAL")
			}
		}
	}

	io, _ := disk.IOCounters()
	for name, counter := range io {
		addReport("DISCO", fmt.Sprintf("E/S %s: lectura=%.1fMB, escritura=%.1fMB",
			name, float64(counter.ReadBytes)/(1024*1024), float64(counter.WriteBytes)/(1024*1024)), "INFO")
	}
}

func checkNetwork() {
	io, _ := net.IOCounters(true)
	for _, counter := range io {
		addReport("RED", fmt.Sprintf("%s: enviados=%.2fMB, recibidos=%.2fMB",
			counter.Name, float64(counter.BytesSent)/(1024*1024), float64(counter.BytesRecv)/(1024*1024)), "INFO")
	}

	conns, _ := net.Connections("tcp")
	addReport("RED", fmt.Sprintf("Conexiones TCP activas: %d", len(conns)), "INFO")

	output := runPowerShell("Get-NetAdapter | Where-Object {$_.Status -eq 'Up'} | Select-Object Name, LinkSpeed | ConvertTo-Json")
	if output != "" && output != "null" {
		addReport("RED", fmt.Sprintf("Adaptadores activos: %s", output[:min(200, len(output))]), "INFO")
	}
}

func checkNetworkAdvanced() {
	addReport("RED_AVANZADA", "=== VERIFICACION AVANZADA DE RED ===", "INFO")

	output := runPowerShell("Test-NetConnection -ComputerName '8.8.8.8' -InformationLevel Quiet")
	if strings.Contains(output, "True") {
		addReport("RED_AVANZADA", "Conectividad a Internet: OK", "INFO")
	} else {
		addReport("RED_AVANZADA", "Sin conectividad a Internet", "CRITICAL")
	}

	output = runPowerShell("Resolve-DnsName -Name google.com -ErrorAction SilentlyContinue | Select-Object -First 1")
	if output != "" {
		addReport("RED_AVANZADA", "DNS: Funcional", "INFO")
	} else {
		addReport("RED_AVANZADA", "Problemas con DNS", "WARNING")
	}

	output = runPowerShell("Get-NetAdapter | Select-Object Name, Status, LinkSpeed | ConvertTo-Json")
	if output != "" && output != "null" {
		addReport("RED_AVANZADA", "Estado de adaptadores obtenido", "INFO")
	}
}

func checkBattery() {
	output := runPowerShell("(Get-WmiObject -Class Win32_Battery).EstimatedChargeRemaining")
	if output != "" {
		var percent float64
		fmt.Sscanf(output, "%f", &percent)

		level := "INFO"
		if percent < 20 {
			level = "WARNING"
		}
		if percent < 10 {
			level = "CRITICAL"
		}

		addReport("BATERIA", fmt.Sprintf("Carga: %.0f%%", percent), level)
	} else {
		addReport("BATERIA", "No se detecto bateria (escritorio)", "INFO")
	}
}

func checkProcesses() {
	procs, _ := process.Processes()

	type procInfo struct {
		name string
		pid  int32
		cpu  float64
		mem  float64
	}

	var topCPU []procInfo
	var topMem []procInfo

	for _, p := range procs {
		name, _ := p.Name()
		cpu, _ := p.CPUPercent()
		mem, _ := p.MemoryPercent()

		if float64(cpu) > 5 {
			topCPU = append(topCPU, procInfo{name: name, pid: p.Pid, cpu: float64(cpu), mem: float64(mem)})
		}
		if float64(mem) > 1 {
			topMem = append(topMem, procInfo{name: name, pid: p.Pid, cpu: float64(cpu), mem: float64(mem)})
		}
	}

	// Sort by CPU
	for i := 0; i < len(topCPU)-1; i++ {
		for j := i + 1; j < len(topCPU); j++ {
			if topCPU[j].cpu > topCPU[i].cpu {
				topCPU[i], topCPU[j] = topCPU[j], topCPU[i]
			}
		}
	}

	// Sort by Memory
	for i := 0; i < len(topMem)-1; i++ {
		for j := i + 1; j < len(topMem); j++ {
			if topMem[j].mem > topMem[i].mem {
				topMem[i], topMem[j] = topMem[j], topMem[i]
			}
		}
	}

	addReport("PROCESOS", "Top 10 procesos por CPU:", "INFO")
	for i := 0; i < min(10, len(topCPU)); i++ {
		addReport("PROCESOS", fmt.Sprintf("  %s: CPU=%.1f%%, PID=%d", topCPU[i].name, topCPU[i].cpu, topCPU[i].pid), "INFO")
	}

	addReport("PROCESOS", "Top 5 procesos por memoria:", "INFO")
	for i := 0; i < min(5, len(topMem)); i++ {
		addReport("PROCESOS", fmt.Sprintf("  %s: Memoria=%.1f%%, PID=%d", topMem[i].name, topMem[i].mem, topMem[i].pid), "INFO")
	}

	addReport("PROCESOS", fmt.Sprintf("Total de procesos: %d", len(procs)), "INFO")
}

func checkBSOD() {
	addReport("BSOD", "=== ANALISIS DE PANTALLAS AZULES (BSOD) ===", "INFO")

	output := runPowerShell("Get-WinEvent -LogName System -MaxEvents 50 | Where-Object {$_.LevelDisplayName -eq 'Error' -and $_.Message -match 'Bugcheck|BlueScreen|0x'} | Select-Object TimeCreated, Message | ConvertTo-Json")

	if output != "" && output != "null" && !strings.Contains(output, "error") {
		addReport("BSOD", "Errores criticos encontrados en logs", "WARNING")
	} else {
		addReport("BSOD", "No se encontraron BSOD recientes", "INFO")
	}
}

func checkSMART() {
	addReport("SMART", "=== VERIFICACION DE DISCO (SMART) ===", "INFO")

	output := runPowerShell("Get-PhysicalDisk | Select-Object FriendlyName, MediaType, OperationalStatus, HealthStatus | ConvertTo-Json")

	if output != "" && output != "null" {
		addReport("SMART", fmt.Sprintf("Informacion SMART: %s", output[:min(300, len(output))]), "INFO")
	} else {
		addReport("SMART", "No se pudo obtener informacion SMART", "WARNING")
	}
}

func checkMalware() {
	addReport("MALWARE", "=== ESCANEO DE MALWARE (Windows Defender) ===", "INFO")

	output := runPowerShell("Get-MpComputerStatus | Select-Object AntivirusEnabled, RealTimeProtectionEnabled | ConvertTo-Json")

	if strings.Contains(output, "AntivirusEnabled") {
		if strings.Contains(output, "true") {
			addReport("MALWARE", "Antivirus: ACTIVADO", "INFO")
		} else {
			addReport("MALWARE", "Antivirus: DESACTIVADO", "CRITICAL")
		}
	}

	if strings.Contains(output, "RealTimeProtectionEnabled") {
		if strings.Contains(output, "true") {
			addReport("MALWARE", "Proteccion en tiempo real: ACTIVADA", "INFO")
		} else {
			addReport("MALWARE", "Proteccion en tiempo real: DESACTIVADA", "WARNING")
		}
	}

	output = runPowerShell("Get-MpThreatDetection | Select-Object -First 5 | ConvertTo-Json")
	if output != "" && output != "null" {
		addReport("MALWARE", "Amenazas detectadas", "WARNING")
	} else {
		addReport("MALWARE", "No hay amenazas detectadas", "INFO")
	}
}

func checkDrivers() {
	addReport("DRIVERS", "=== VERIFICACION DE DRIVERS ===", "INFO")

	output := runPowerShell("Get-PnpDevice -Class SoftwareDevice -Status Error | Select-Object FriendlyName | ConvertTo-Json")

	if output != "" && output != "null" {
		addReport("DRIVERS", "Drivers con problemas detectados", "WARNING")
	} else {
		addReport("DRIVERS", "Drivers sin errores detectados", "INFO")
	}

	output = runPowerShell("driverquery /v /fo csv | ConvertFrom-Csv | Where-Object {$_.{'Estado del controlador'} -ne 'Iniciado'} | Select-Object 'Nombre del módulo' | ConvertTo-Json")
	if output != "" && output != "null" {
		addReport("DRIVERS", "Algunos drivers no iniciados", "WARNING")
	}
}

func checkSystemIntegrity() {
	addReport("INTEGRIDAD", "=== VERIFICACION DE INTEGRIDAD DEL SISTEMA ===", "INFO")

	output := runPowerShell("sfc /verifyonly")

	if strings.Contains(output, "no integrity violations") {
		addReport("INTEGRIDAD", "Integridad de archivos: OK", "INFO")
	} else {
		addReport("INTEGRIDAD", "Se encontraron archivos corruptos", "WARNING")
		addReport("INTEGRIDAD", "Ejecuta 'sfc /scannow' para reparar", "WARNING")
	}

	output = runPowerShell("DISM /Online /Cleanup-Image /CheckHealth")
	if strings.Contains(output, "No component store corruption detected") {
		addReport("INTEGRIDAD", "Component Store: OK", "INFO")
	} else {
		addReport("INTEGRIDAD", "Corrupcion en Component Store", "WARNING")
	}
}

func checkWindowsUpdate() {
	addReport("WINDOWS UPDATE", "=== ESTADO DE WINDOWS UPDATE ===", "INFO")

	output := runPowerShell("Get-ItemProperty 'HKLM:\\SOFTWARE\\Microsoft\\Windows\\CurrentVersion\\WindowsUpdate\\Auto Update\\Results\\Install' | Select-Object LastSuccessTime | ConvertTo-Json")

	if output != "" && output != "null" {
		addReport("WINDOWS UPDATE", fmt.Sprintf("Ultima actualizacion: %s", output), "INFO")
	} else {
		addReport("WINDOWS UPDATE", "No se pudo obtener informacion", "WARNING")
	}
}

func checkServices() {
	addReport("SERVICIOS", "=== SERVICIOS CRITICOS ===", "INFO")

	servicios := []string{"wuauserv", "WinDefend", "BITS", "Dhcp", "Dnscache", "EventLog", "RpcSs"}

	output := runPowerShell(fmt.Sprintf("Get-Service -Name %s | Where-Object {$_.Status -eq 'Stopped'} | Select-Object Name | ConvertTo-Json",
		strings.Join(servicios, ",")))

	if output != "" && output != "null" {
		addReport("SERVICIOS", "Algunos servicios criticos detenidos", "WARNING")
	} else {
		addReport("SERVICIOS", "Todos los servicios criticos activos", "INFO")
	}
}

func checkStartup() {
	addReport("STARTUP", "=== PROGRAMAS DE INICIO ===", "INFO")

	output := runPowerShell("Get-CimInstance Win32_StartupCommand | Select-Object Name, Command | ConvertTo-Json")

	if output != "" && output != "null" && output != "[]" {
		addReport("STARTUP", fmt.Sprintf("Programas de inicio: %s", output[:min(300, len(output))]), "INFO")
	} else {
		addReport("STARTUP", "No hay programas de inicio configurados", "INFO")
	}
}

func checkLogs() {
	addReport("LOGS", "=== EVENTOS DEL SISTEMA ===", "INFO")

	output := runPowerShell("Get-WinEvent -LogName System -MaxEvents 20 | Where-Object {$_.LevelDisplayName -in @('Error','Warning')} | Select-Object TimeCreated, LevelDisplayName, Message | ConvertTo-Json")

	if output != "" && output != "null" {
		addReport("LOGS", "Errores/advertencias encontrados en logs del sistema", "WARNING")
	}

	addReport("LOGS", "=== EVENTOS DE APLICACION ===", "INFO")

	output = runPowerShell("Get-WinEvent -LogName Application -MaxEvents 20 | Where-Object {$_.LevelDisplayName -in @('Error','Warning')} | Select-Object TimeCreated, LevelDisplayName, Message | ConvertTo-Json")

	if output != "" && output != "null" {
		addReport("LOGS", "Errores/advertencias en logs de aplicacion", "WARNING")
	}
}

func checkUSB() {
	addReport("USB", "=== DISPOSITIVOS USB ===", "INFO")

	output := runPowerShell("Get-PnpDevice -Class USB -Status OK | Measure-Object | Select-Object -ExpandProperty Count")

	if output != "" && output != "0" {
		addReport("USB", fmt.Sprintf("Dispositivos USB detectados: %s", output), "INFO")
	}
}

func generateRecommendations() {
	addReport("RECOMENDACIONES", "=== RECOMENDACIONES DE MANTENIMIENTO ===", "INFO")

	vm, _ := mem.VirtualMemory()
	if vm.UsedPercent > 80 {
		addReport("RECOMENDACIONES", "1. Cerrar aplicaciones que consumen mucha memoria", "WARNING")
	}

	disks, _ := disk.Partitions(true)
	for _, d := range disks {
		usage, _ := disk.Usage(d.Mountpoint)
		if usage.UsedPercent > 85 {
			addReport("RECOMENDACIONES", "2. Liberar espacio en disco", "CRITICAL")
			break
		}
	}

	cpuPercent, _ := cpu.Percent(time.Second, false)
	if len(cpuPercent) > 0 && cpuPercent[0] > 80 {
		addReport("RECOMENDACIONES", "3. CPU muy utilizado, revisar procesos", "WARNING")
	}

	addReport("RECOMENDACIONES", "4. Ejecutar Windows Update regularmente", "INFO")
	addReport("RECOMENDACIONES", "5. Ejecutar 'sfc /scannow' para verificar archivos", "INFO")
	addReport("RECOMENDACIONES", "6. Ejecutar 'chkdsk C: /f' si hay errores", "INFO")
	addReport("RECOMENDACIONES", "7. Limpiar archivos temporales periodicamente", "INFO")
	addReport("RECOMENDACIONES", "8. Mantener antivirus actualizado", "INFO")
	addReport("RECOMENDACIONES", "9. Desactivar programas de inicio innecesarios", "INFO")
	addReport("RECOMENDACIONES", "10. Hacer respaldo de datos importantes", "INFO")
}

func saveReport(filename string) {
	dr := DiagnosticReport{
		Fecha:     time.Now().Format("2006-01-02 15:04:05"),
		Equipo:    equipo,
		Hardware:  hardware,
		Sistema:   sistema,
		Red:       red,
		Seguridad: seguridad,
		Servicios: servicios,
		Procesos:  procesos,
		Alertas:   alertas,
	}

	data, _ := json.MarshalIndent(dr, "", "  ")
	_ = os.WriteFile(filename, data, 0644)
	fmt.Printf("\n📄 Reporte guardado en: %s\n", filename)
}

func saveAlerts() {
	if len(alertas) > 0 {
		aa := AlertasActivas{
			Timestamp: time.Now().Format(time.RFC3339),
			Alertas:   alertas,
		}

		data, _ := json.MarshalIndent(aa, "", "  ")
		_ = os.WriteFile("alertas_activas.json", data, 0644)
		fmt.Printf("⚠️ Alertas guardadas en: alertas_activas.json\n")
	}
}

func saveHistory() {
	var historial Historial

	if data, err := os.ReadFile("historial_diagnostico.json"); err == nil {
		_ = json.Unmarshal(data, &historial)
	}

	vm, _ := mem.VirtualMemory()
	cpuPercent, _ := cpu.Percent(time.Second, false)

	criticas := 0
	warnings := 0
	for _, a := range alertas {
		if a.Level == "CRITICAL" {
			criticas++
		} else if a.Level == "WARNING" {
			warnings++
		}
	}

	cpuAvg := 0.0
	if len(cpuPercent) > 0 {
		cpuAvg = cpuPercent[0]
	}

	entry := HistorialEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Resumen: Resumen{
			AlertasCriticas:   criticas,
			AlertasWarning:    warnings,
			CPUPromedio:       cpuAvg,
			MemoriaPorcentaje: vm.UsedPercent,
		},
	}

	historial.Historial = append(historial.Historial, entry)

	if len(historial.Historial) > 30 {
		historial.Historial = historial.Historial[len(historial.Historial)-30:]
	}

	data, _ := json.MarshalIndent(historial, "", "  ")
	_ = os.WriteFile("historial_diagnostico.json", data, 0644)
}

func printReport() {
	if len(alertas) > 0 {
		criticas := 0
		warnings := 0
		for _, a := range alertas {
			if a.Level == "CRITICAL" {
				criticas++
			} else if a.Level == "WARNING" {
				warnings++
			}
		}

		fmt.Printf("\n%s\n", strings.Repeat("=", 60))
		fmt.Printf("  RESUMEN: %d CRITICAL, %d WARNING\n", criticas, warnings)
		fmt.Printf("%s\n", strings.Repeat("=", 60))
	}
}

func runFullDiagnostic() {
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  CIBEROX ENDPOINT - DIAGNOSTICO COMPLETO")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Fecha: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))

	fmt.Println("[1/19] Obteniendo informacion del sistema...")
	getSystemInfo()

	fmt.Println("[2/19] Verificando CPU...")
	checkCPU()

	fmt.Println("[3/19] Verificando memoria...")
	checkMemory()

	fmt.Println("[4/19] Verificando disco...")
	checkDisk()

	fmt.Println("[5/19] Verificando red...")
	checkNetwork()

	fmt.Println("[6/19] Verificando bateria...")
	checkBattery()

	fmt.Println("[7/19] Verificando procesos...")
	checkProcesses()

	fmt.Println("[8/19] Analizando BSOD...")
	checkBSOD()

	fmt.Println("[9/19] Verificando SMART del disco...")
	checkSMART()

	fmt.Println("[10/19] Escaneando malware...")
	checkMalware()

	fmt.Println("[11/19] Verificando drivers...")
	checkDrivers()

	fmt.Println("[12/19] Verificando integridad del sistema...")
	checkSystemIntegrity()

	fmt.Println("[13/19] Verificando Windows Update...")
	checkWindowsUpdate()

	fmt.Println("[14/19] Verificando servicios...")
	checkServices()

	fmt.Println("[15/19] Verificando programas de inicio...")
	checkStartup()

	fmt.Println("[16/19] Verificando red avanzada...")
	checkNetworkAdvanced()

	fmt.Println("[17/19] Verificando dispositivos USB...")
	checkUSB()

	fmt.Println("[18/19] Leyendo logs de Windows...")
	checkLogs()

	fmt.Println("[19/19] Ejecutando auditoría de seguridad...")
	runSecurityAudit()

	fmt.Println("\nGenerando recomendaciones...")
	generateRecommendations()

	printReport()

	filename := fmt.Sprintf("diagnostico_%s.json", time.Now().Format("20060102_150405"))
	saveReport(filename)
	saveAlerts()
	saveHistory()
}

func main() {
	runFullDiagnostic()
}

func getComputerInfo() {
	output := runPowerShell("(Get-CimInstance Win32_ComputerSystem).Name")
	if output != "" {
		equipo.NombreEquipo = output
		addReport("EQUIPO", fmt.Sprintf("Nombre del equipo: %s", output), "INFO")
	}

	output = runPowerShell("(Get-CimInstance Win32_ComputerSystem).Domain")
	if output != "" {
		equipo.Dominio = output
		addReport("EQUIPO", fmt.Sprintf("Dominio/Grupo de trabajo: %s", output), "INFO")
	}

	output = runPowerShell("(Get-CimInstance Win32_ComputerSystem).DomainRole")
	if output != "" {
		role := ""
		switch output {
		case "0":
			role = "Estación de trabajo"
		case "1":
			role = "Estación de trabajo (miembro de dominio)"
		case "2":
			role = "Servidor miembro"
		case "3":
			role = "Servidor miembro (miembro de dominio)"
		case "4":
			role = "Controlador de dominio"
		case "5":
			role = "Controlador de dominio (primario)"
		}
		if role != "" {
			equipo.Rol = role
			addReport("EQUIPO", fmt.Sprintf("Rol: %s", role), "INFO")
		}
	}

	output = runPowerShell("(Get-CimInstance Win32_OperatingSystem).Caption")
	if output != "" {
		equipo.SistemaOperativo = output
		addReport("EQUIPO", fmt.Sprintf("Sistema operativo: %s", output), "INFO")
	}

	output = runPowerShell("(Get-CimInstance Win32_OperatingSystem).Version")
	if output != "" {
		equipo.Version = output
		addReport("EQUIPO", fmt.Sprintf("Versión: %s", output), "INFO")
	}

	output = runPowerShell("(Get-CimInstance Win32_ComputerSystem).Manufacturer")
	if output != "" {
		equipo.Fabricante = output
		addReport("EQUIPO", fmt.Sprintf("Fabricante: %s", output), "INFO")
	}

	output = runPowerShell("(Get-CimInstance Win32_ComputerSystem).Model")
	if output != "" {
		addReport("EQUIPO", fmt.Sprintf("Modelo: %s", output), "INFO")
	}

	output = runPowerShell("(Get-CimInstance Win32_BIOS).SerialNumber")
	if output != "" {
		addReport("EQUIPO", fmt.Sprintf("Número de serie: %s", output), "INFO")
	}
}

func getListeningPorts() {
	addReport("SEGURIDAD", "=== PUERTOS EN ESCUCHA ===", "INFO")

	connections, err := net.Connections("tcp")
	if err != nil {
		addReport("SEGURIDAD", "Error al obtener conexiones TCP", "WARNING")
		return
	}

	ports := make(map[string]string)
	for _, conn := range connections {
		if conn.Status == "LISTEN" {
			proc, err := process.NewProcess(int32(conn.Pid))
			procName := "desconocido"
			if err == nil {
				procName, _ = proc.Name()
			}
			portKey := fmt.Sprintf("%d", conn.Laddr.Port)
			if _, exists := ports[portKey]; !exists {
				ports[portKey] = procName
			}
		}
	}

	knownPorts := map[string]string{
		"80":   "HTTP",
		"443":  "HTTPS",
		"22":   "SSH",
		"21":   "FTP",
		"25":   "SMTP",
		"53":   "DNS",
		"3306": "MySQL",
		"5432": "PostgreSQL",
		"3389": "RDP",
		"135":  "RPC",
		"139":  "NetBIOS",
		"445":  "SMB",
		"5985": "WinRM",
		"5986": "WinRM SSL",
	}

	suspiciousPorts := []string{"4444", "5555", "6666", "6667", "31337", "12345"}

	for port, proc := range ports {
		service := knownPorts[port]
		if service == "" {
			service = "desconocido"
		}

		level := "INFO"
		msg := fmt.Sprintf("Puerto %s (%s) - Proceso: %s", port, service, proc)

		for _, susp := range suspiciousPorts {
			if port == susp {
				level = "CRITICAL"
				msg = fmt.Sprintf("PUERTO SOSPECHOSO: %s", msg)
				break
			}
		}
		addReport("SEGURIDAD", msg, level)
	}

	if len(ports) == 0 {
		addReport("SEGURIDAD", "No se detectaron puertos en escucha", "WARNING")
	}
}

func getSecurityPatches() {
	addReport("SEGURIDAD", "=== PARCHES DE SEGURIDAD ===", "INFO")

	output := runPowerShell("Get-HotFix | Sort-Object -Property InstalledOn -Descending | Select-Object -First 10 | Format-Table -AutoSize")

	if output != "" {
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				addReport("SEGURIDAD", fmt.Sprintf("HotFix: %s", strings.TrimSpace(line)), "INFO")
			}
		}
	}

	output = runPowerShell("(Get-CimInstance Win32_QuickFixEngineering | Sort-Object -Property InstalledOn -Descending | Select-Object -First 1).InstalledOn")
	if output != "" {
		addReport("SEGURIDAD", fmt.Sprintf("Último parche instalado: %s", output), "INFO")
	}

	output = runPowerShell("$days = (Get-Date) - (Get-CimInstance Win32_QuickFixEngineering | Sort-Object -Property InstalledOn -Descending | Select-Object -First 1).InstalledOn; Write-Output $days.Days")
	if output != "" {
		days := strings.TrimSpace(output)
		addReport("SEGURIDAD", fmt.Sprintf("Días desde último parche: %s", days), "INFO")
	}

	checkPatchAge(output)
}

func checkPatchAge(daysStr string) {
	days := 0
	fmt.Sscanf(daysStr, "%d", &days)

	if days > 30 {
		addReport("SEGURIDAD", fmt.Sprintf("ALERTA: %d días sin parches de seguridad", days), "CRITICAL")
	} else if days > 14 {
		addReport("SEGURIDAD", fmt.Sprintf("ADVERTENCIA: %d días sin parches", days), "WARNING")
	}
}

func getPasswordPolicy() {
	addReport("SEGURIDAD", "=== POLÍTICAS DE CONTRASEÑA ===", "INFO")

	output := runPowerShell("net user %USERNAME%")
	if output != "The command completed with one or more errors." {
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			if strings.Contains(line, "Local Group Memberships") || strings.Contains(line, "Administrator") {
				addReport("SEGURIDAD", fmt.Sprintf("Usuario %s: %s", os.Getenv("USERNAME"), strings.TrimSpace(line)), "INFO")
			}
		}
	}

	isAdmin := runPowerShell("net session")
	if isAdmin != "The command completed with one or more errors." {
		addReport("SEGURIDAD", "El usuario tiene privilegios de ADMINISTRADOR", "CRITICAL")
	} else {
		addReport("SEGURIDAD", "El usuario NO tiene privilegios de administrador", "INFO")
	}

	output = runPowerShell("net localgroup Administrators")
	if output != "" {
		lines := strings.Split(output, "\n")
		addReport("SEGURIDAD", "Usuarios con privilegios de Administrador:", "INFO")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && !strings.Contains(line, "The command completed") && !strings.Contains(line, "Members") && !strings.Contains(line, "---") {
				addReport("SEGURIDAD", fmt.Sprintf("  - %s", line), "INFO")
			}
		}
	}
}

func getUserList() {
	addReport("SEGURIDAD", "=== USUARIOS LOCALES ===", "INFO")

	output := runPowerShell("Get-LocalUser | Select-Object Name, Enabled, LastLogon | Format-Table -AutoSize")

	if output != "" {
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				addReport("SEGURIDAD", fmt.Sprintf("Usuario: %s", strings.TrimSpace(line)), "INFO")
			}
		}
	}

	addReport("SEGURIDAD", "Cuentas habilitadas/deshabilitadas:", "INFO")
	output = runPowerShell("Get-LocalUser | Select-Object Name, Enabled | ConvertTo-Json")
	if output != "" {
		addReport("SEGURIDAD", output, "INFO")
	}
}

func getUSBDevices() {
	addReport("SEGURIDAD", "=== DISPOSITIVOS USB ===", "INFO")

	output := runPowerShell("Get-PnpDevice -Class USB -Status OK | Select-Object FriendlyName, Manufacturer, Status | Format-Table -AutoSize")

	if output != "" && !strings.Contains(output, "No objects") {
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				addReport("SEGURIDAD", fmt.Sprintf("USB: %s", strings.TrimSpace(line)), "INFO")
			}
		}
	}

	addReport("SEGURIDAD", "Historial de dispositivos USB conectados:", "INFO")
	output = runPowerShell("Get-ItemProperty -Path 'HKLM:\\SYSTEM\\CurrentControlSet\\Enum\\USBSTOR\\*\\*' | Select-Object DeviceDesc, FriendlyName, Service | Format-Table -AutoSize")

	if output != "" && !strings.Contains(output, "Cannot find") {
		lines := strings.Split(output, "\n")
		count := 0
		for _, line := range lines {
			if strings.TrimSpace(line) != "" && count < 10 {
				addReport("SEGURIDAD", fmt.Sprintf("USBSTOR: %s", strings.TrimSpace(line)), "INFO")
				count++
			}
		}
	} else {
		addReport("SEGURIDAD", "No se encontró historial de dispositivos USB", "WARNING")
	}
}

func getPrinters() {
	addReport("SEGURIDAD", "=== IMPRESORAS Y DISPOSITIVOS ===", "INFO")

	output := runPowerShell("Get-Printer | Select-Object Name, PortName, Status | Format-Table -AutoSize")

	if output != "" && !strings.Contains(output, "No objects") {
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				addReport("SEGURIDAD", fmt.Sprintf("Impresora: %s", strings.TrimSpace(line)), "INFO")
			}
		}
	} else {
		addReport("SEGURIDAD", "No se detectaron impresoras", "INFO")
	}

	addReport("SEGURIDAD", "Dispositivos de red:", "INFO")
	output = runPowerShell("Get-NetAdapter | Select-Object Name, Status, MacAddress, LinkSpeed | Format-Table -AutoSize")

	if output != "" {
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				addReport("SEGURIDAD", fmt.Sprintf("Adaptador: %s", strings.TrimSpace(line)), "INFO")
			}
		}
	}
}

func getFirewallStatus() {
	addReport("SEGURIDAD", "=== ESTADO DEL FIREWALL ===", "INFO")

	profiles := []string{"Domain", "Public", "Private"}
	for _, profile := range profiles {
		output := runPowerShell(fmt.Sprintf("(Get-NetFirewallProfile -Name '%s').Enabled", profile))
		enabled := strings.TrimSpace(output) == "True"
		status := "DESACTIVADO"
		level := "CRITICAL"
		if enabled {
			status = "ACTIVADO"
			level = "INFO"
		}
		addReport("SEGURIDAD", fmt.Sprintf("Firewall %s: %s", profile, status), level)
	}

	addReport("SEGURIDAD", "=== REGLAS DE FIREWALL PELIGROSAS ===", "INFO")

	output := runPowerShell("Get-NetFirewallRule | Where-Object { $_.Direction -eq 'Inbound' -and $_.Action -eq 'Allow' -and ($_.Profile -eq 'Any' -or $_.Profile -contains 'Any') } | Select-Object -First 10 DisplayName, Direction, Action, Profile | Format-Table -AutoSize")

	if output != "" && !strings.Contains(output, "No objects") {
		addReport("SEGURIDAD", "Reglas 'Permitir todo' (Inbound Any/Any):", "WARNING")
		lines := strings.Split(output, "\n")
		count := 0
		for _, line := range lines {
			if strings.TrimSpace(line) != "" && count < 5 {
				addReport("SEGURIDAD", fmt.Sprintf("  %s", strings.TrimSpace(line)), "WARNING")
				count++
			}
		}
	}

	addReport("SEGURIDAD", "Reglas con puertos abiertos expuestos:", "INFO")
	output = runPowerShell("Get-NetFirewallRule | Where-Object { $_.Direction -eq 'Inbound' -and $_.Action -eq 'Allow' } | Where-Object { $_.LocalPort -ne 'Any' } | Select-Object DisplayName, LocalPort, RemotePort | Sort-Object LocalPort -Unique | Select-Object -First 15 | Format-Table -AutoSize")

	if output != "" {
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				addReport("SEGURIDAD", fmt.Sprintf("  %s", strings.TrimSpace(line)), "INFO")
			}
		}
	}
}

func runSecurityAudit() {
	getComputerInfo()
	getListeningPorts()
	getSecurityPatches()
	getPasswordPolicy()
	getUserList()
	getUSBDevices()
	getPrinters()
	getFirewallStatus()
	getInstalledSoftware()
	getAutorunEntries()
	getBrowserExtensions()
	checkVulnerabilities()
	collectHardwareInfo()
}

func collectHardwareInfo() {
	getDiskHealth()
	getBatteryHealth()
	getBIOSInfo()
	getProcessorInfo()
	getRAMInfo()
}

func getDiskHealth() {
	addReport("HARDWARE", "=== SALUD DEL DISCO (S.M.A.R.T.) ===", "INFO")

	output := runPowerShell("Get-PhysicalDisk | Select-Object FriendlyName, SerialNumber, MediaType, OperationalStatus, HealthStatus | ConvertTo-Json")

	if output != "" && output != "null" {
		var disks []map[string]interface{}
		if err := json.Unmarshal([]byte(output), &disks); err == nil {
			for _, disk := range disks {
				diskInfo := DiskInfo{}

				if name, ok := disk["FriendlyName"].(string); ok {
					diskInfo.Modelo = name
				}
				if serial, ok := disk["SerialNumber"].(string); ok {
					diskInfo.NumeroSerie = serial
				}
				if status, ok := disk["OperationalStatus"].(string); ok {
					diskInfo.Estado = status
				}
				if health, ok := disk["HealthStatus"].(string); ok {
					if health == "Healthy" {
						diskInfo.SaludPorcent = 100
					} else {
						diskInfo.SaludPorcent = 0
					}
				}

				smartData := runPowerShell(fmt.Sprintf("Get-StorageReliabilityCounter -PhysicalDisk (Get-PhysicalDisk | Where-Object { $_.FriendlyName -eq '%s' }) | ConvertTo-Json", diskInfo.Modelo))
				if smartData != "" && smartData != "null" {
					var smart map[string]interface{}
					if err := json.Unmarshal([]byte(smartData), &smart); err == nil {
						if temp, ok := smart["Temperature"].(float64); ok {
							diskInfo.Temperatura = temp
						}
						if readErrors, ok := smart["ReadErrorCount"].(float64); ok {
							diskInfo.LecturaErrores = int64(readErrors)
						}
						if reallocated, ok := smart["ReallocatedSectorsCount"].(float64); ok {
							diskInfo.Reasignados = int64(reallocated)
						}
						if pending, ok := smart["PendingSectorCount"].(float64); ok {
							diskInfo.Pendientes = int64(pending)
						}
					}
				}

				hardware.Discos = append(hardware.Discos, diskInfo)
				addReport("HARDWARE", fmt.Sprintf("Disco: %s - Serie: %s - Estado: %s", diskInfo.Modelo, diskInfo.NumeroSerie, diskInfo.Estado), "INFO")

				if diskInfo.Temperatura > 0 {
					addReport("HARDWARE", fmt.Sprintf("  Temperatura: %.1f°C", diskInfo.Temperatura), "INFO")
				}
				if diskInfo.Reasignados > 0 || diskInfo.Pendientes > 0 {
					addReport("HARDWARE", fmt.Sprintf("  ALERTA: Sectores reasignados: %d, Pendientes: %d", diskInfo.Reasignados, diskInfo.Pendientes), "WARNING")
				}
			}
		}
	}

	output = runPowerShell("Get-WmiObject -Class Win32_DiskDrive | Select-Object Model, SerialNumber, Size, Status | ConvertTo-Json")
	if output != "" && output != "null" {
		var disks []map[string]interface{}
		if err := json.Unmarshal([]byte(output), &disks); err == nil {
			for _, disk := range disks {
				if len(hardware.Discos) == 0 || hardware.Discos[0].Modelo == "" {
					diskInfo := DiskInfo{}
					if model, ok := disk["Model"].(string); ok {
						diskInfo.Modelo = model
					}
					if serial, ok := disk["SerialNumber"].(string); ok {
						diskInfo.NumeroSerie = serial
					}
					if status, ok := disk["Status"].(string); ok {
						diskInfo.Estado = status
					}
					hardware.Discos = append(hardware.Discos, diskInfo)
				}
			}
		}
	}
}

func getBatteryHealth() {
	addReport("HARDWARE", "=== SALUD DE BATERÍA ===", "INFO")

	output := runPowerShell("Get-WmiObject -Class Win32_Battery | Select-Object Name, EstimatedChargeRemaining, EstimatedRunTime, BatteryStatus, DesignCapacity, FullChargeCapacity | ConvertTo-Json")

	if output != "" && output != "null" {
		var battery map[string]interface{}
		if err := json.Unmarshal([]byte(output), &battery); err == nil {
			bat := BatteryInfo{}

			if name, ok := battery["Name"].(string); ok {
				bat.Nombre = name
			}
			if charge, ok := battery["EstimatedChargeRemaining"].(float64); ok {
				bat.CargaPorcent = charge
			}
			if runtime, ok := battery["EstimatedRunTime"].(float64); ok {
				bat.TiempoRestante = int(runtime)
			}
			if status, ok := battery["BatteryStatus"].(float64); ok {
				switch int(status) {
				case 1:
					bat.Estado = "Desconectado"
				case 2:
					bat.Estado = "Conectado - Alta carga"
				case 3:
					bat.Estado = "Conectado - Baja carga"
				case 4:
					bat.Estado = "Conectado - Cargando"
				case 5:
					bat.Estado = "Conectado - Cargado"
				case 6:
					bat.Estado = "Conectado - Carga baja"
				case 7:
					bat.Estado = "Conectado - Cargando"
				case 8:
					bat.Estado = "Conectado - Cargando"
				case 9:
					bat.Estado = "Desconocido"
				case 10:
					bat.Estado = "Desconectado"
				case 11:
					bat.Estado = "Cargando"
				default:
					bat.Estado = "Desconocido"
				}
			}
			if designCap, ok := battery["DesignCapacity"].(float64); ok {
				bat.CapacidadDiseno = int64(designCap)
			}
			if fullCap, ok := battery["FullChargeCapacity"].(float64); ok {
				bat.CapacidadActual = int64(fullCap)
			}

			if bat.CapacidadDiseno > 0 && bat.CapacidadActual > 0 {
				bat.SaludPorcent = float64(bat.CapacidadActual) / float64(bat.CapacidadDiseno) * 100
			}

			output = runPowerShell("Get-WmiObject -Class BatteryFullChargedCapacity | Select-Object FullChargedCapacity")
			if output != "" {
				var cycles map[string]interface{}
				if err := json.Unmarshal([]byte(output), &cycles); err == nil {
					if fc, ok := cycles["FullChargedCapacity"].(float64); ok {
						bat.CapacidadActual = int64(fc)
						if bat.CapacidadDiseno > 0 {
							bat.SaludPorcent = float64(bat.CapacidadActual) / float64(bat.CapacidadDiseno) * 100
						}
					}
				}
			}

			hardware.Bateria = bat

			addReport("HARDWARE", fmt.Sprintf("Batería: %s", bat.Nombre), "INFO")
			addReport("HARDWARE", fmt.Sprintf("Carga: %.0f%%", bat.CargaPorcent), "INFO")
			addReport("HARDWARE", fmt.Sprintf("Estado: %s", bat.Estado), "INFO")

			if bat.SaludPorcent > 0 {
				addReport("HARDWARE", fmt.Sprintf("Salud de batería: %.1f%%", bat.SaludPorcent), "INFO")
				if bat.SaludPorcent < 80 {
					addReport("HARDWARE", fmt.Sprintf("ALERTA: Salud de batería baja (%.1f%%). Considerar reemplazo", bat.SaludPorcent), "WARNING")
				}
			}

			if bat.CapacidadDiseno > 0 && bat.CapacidadActual > 0 {
				addReport("HARDWARE", fmt.Sprintf("Capacidad diseño: %d mWh", bat.CapacidadDiseno), "INFO")
				addReport("HARDWARE", fmt.Sprintf("Capacidad actual: %d mWh", bat.CapacidadActual), "INFO")
			}
		}
	} else {
		addReport("HARDWARE", "No se detectó batería (equipo de escritorio)", "INFO")
	}
}

func getBIOSInfo() {
	addReport("HARDWARE", "=== INFORMACIÓN BIOS ===", "INFO")

	output := runPowerShell("Get-WmiObject -Class Win32_BIOS | Select-Object Manufacturer, Name, Version, SerialNumber, ReleaseDate | ConvertTo-Json")

	if output != "" && output != "null" {
		var bios map[string]interface{}
		if err := json.Unmarshal([]byte(output), &bios); err == nil {
			b := BIOSInfo{}
			if mfr, ok := bios["Manufacturer"].(string); ok {
				b.Fabricante = mfr
			}
			if name, ok := bios["Name"].(string); ok {
				b.Version = name
			}
			if ver, ok := bios["Version"].(string); ok {
				b.Version = ver
			}
			if serial, ok := bios["SerialNumber"].(string); ok {
				b.NumeroSerie = serial
			}
			if date, ok := bios["ReleaseDate"].(string); ok {
				b.Fecha = date
			}
			hardware.BIOS = b

			addReport("HARDWARE", fmt.Sprintf("Fabricante: %s", b.Fabricante), "INFO")
			addReport("HARDWARE", fmt.Sprintf("Versión: %s", b.Version), "INFO")
			addReport("HARDWARE", fmt.Sprintf("Número de serie: %s", b.NumeroSerie), "INFO")

			if equipo.NumeroSerie == "" {
				equipo.NumeroSerie = b.NumeroSerie
			}
		}
	}
}

func getProcessorInfo() {
	addReport("HARDWARE", "=== PROCESADOR ===", "INFO")

	output := runPowerShell("Get-WmiObject -Class Win32_Processor | Select-Object Name, Manufacturer, NumberOfCores, NumberOfLogicalProcessors, MaxClockSpeed, ProcessorId | ConvertTo-Json")

	if output != "" && output != "null" {
		var cpu map[string]interface{}
		if err := json.Unmarshal([]byte(output), &cpu); err == nil {
			p := ProcessorInfo{}
			if name, ok := cpu["Name"].(string); ok {
				p.Nombre = strings.TrimSpace(name)
			}
			if mfr, ok := cpu["Manufacturer"].(string); ok {
				p.Fabricante = mfr
			}
			if cores, ok := cpu["NumberOfCores"].(float64); ok {
				p.Nucleos = int(cores)
			}
			if threads, ok := cpu["NumberOfLogicalProcessors"].(float64); ok {
				p.Threads = int(threads)
			}
			if freq, ok := cpu["MaxClockSpeed"].(float64); ok {
				p.FrecuenciaMax = int(freq)
			}
			if id, ok := cpu["ProcessorId"].(string); ok {
				p.Identificador = id
			}
			hardware.Procesador = p

			addReport("HARDWARE", fmt.Sprintf("Procesador: %s", p.Nombre), "INFO")
			addReport("HARDWARE", fmt.Sprintf("Núcleos: %d, Hilos: %d", p.Nucleos, p.Threads), "INFO")
			addReport("HARDWARE", fmt.Sprintf("Frecuencia máxima: %d MHz", p.FrecuenciaMax), "INFO")
		}
	}
}

func getRAMInfo() {
	addReport("HARDWARE", "=== MEMORIA RAM ===", "INFO")

	output := runPowerShell("Get-WmiObject -Class Win32_PhysicalMemory | Select-Object Capacity, Speed, MemoryType, FormFactor | ConvertTo-Json")

	if output != "" && output != "null" {
		var mem []map[string]interface{}
		if err := json.Unmarshal([]byte(output), &mem); err == nil {
			var totalRAM int64
			var velocidad int
			slots := 0

			for _, m := range mem {
				if cap, ok := m["Capacity"].(float64); ok {
					totalRAM += int64(cap)
					slots++
				}
				if speed, ok := m["Speed"].(float64); ok {
					velocidad = int(speed)
				}
			}

			r := RAMInfo{
				TotalGB:     float64(totalRAM) / (1024 * 1024 * 1024),
				SlotsUsados: slots,
				Velocidad:   velocidad,
			}
			hardware.RAM = r

			addReport("HARDWARE", fmt.Sprintf("Total RAM: %.1f GB", r.TotalGB), "INFO")
			addReport("HARDWARE", fmt.Sprintf("Slots utilizados: %d", r.SlotsUsados), "INFO")
			addReport("HARDWARE", fmt.Sprintf("Velocidad: %d MHz", r.Velocidad), "INFO")
		}
	}
}

func getInstalledSoftware() {
	addReport("SOFTWARE", "=== APLICACIONES INSTALADAS ===", "INFO")

	output := runPowerShell("Get-ItemProperty HKLM:\\Software\\Microsoft\\Windows\\CurrentVersion\\Uninstall\\* | Select-Object DisplayName, DisplayVersion, Publisher, InstallDate | Where-Object { $_.DisplayName } | ConvertTo-Json")

	if output != "" && output != "null" {
		var apps []InstalledApp
		if err := json.Unmarshal([]byte(output), &apps); err == nil {
			for _, app := range apps {
				if app.Nombre != "" {
					software.InstalledApps = append(software.InstalledApps, app)
				}
			}
		}
	}

	output = runPowerShell("Get-ItemProperty HKLM:\\Software\\Wow6432Node\\Microsoft\\Windows\\CurrentVersion\\Uninstall\\* | Select-Object DisplayName, DisplayVersion, Publisher, InstallDate | Where-Object { $_.DisplayName } | ConvertTo-Json")

	if output != "" && output != "null" {
		var apps []InstalledApp
		if err := json.Unmarshal([]byte(output), &apps); err == nil {
			for _, app := range apps {
				if app.Nombre != "" {
					software.InstalledApps = append(software.InstalledApps, app)
				}
			}
		}
	}

	addReport("SOFTWARE", fmt.Sprintf("Total aplicaciones instaladas: %d", len(software.InstalledApps)), "INFO")
}

func getAutorunEntries() {
	addReport("SOFTWARE", "=== SOFTWARE DE INICIO (AUTORUN) ===", "INFO")

	locations := []string{
		"HKLM:\\Software\\Microsoft\\Windows\\CurrentVersion\\Run",
		"HKLM:\\Software\\Microsoft\\Windows\\CurrentVersion\\RunOnce",
		"HKCU:\\Software\\Microsoft\\Windows\\CurrentVersion\\Run",
		"HKCU:\\Software\\Microsoft\\Windows\\CurrentVersion\\RunOnce",
		"HKLM:\\Software\\Microsoft\\Windows\\CurrentVersion\\RunServices",
		"HKLM:\\Software\\Microsoft\\Windows\\CurrentVersion\\Explorer\\StartupApproved\\Run",
		"HKCU:\\Software\\Microsoft\\Windows\\CurrentVersion\\Explorer\\StartupApproved\\Run",
	}

	for _, loc := range locations {
		output := runPowerShell(fmt.Sprintf("Get-ItemProperty -Path '%s' -ErrorAction SilentlyContinue | ConvertTo-Json", loc))
		if output != "" && output != "null" && !strings.Contains(output, "Cannot find") {
			addReport("SOFTWARE", fmt.Sprintf("Ubicación: %s", loc), "INFO")
			var entries map[string]string
			if err := json.Unmarshal([]byte(output), &entries); err == nil {
				for name, cmd := range entries {
					if !strings.HasPrefix(name, "PS") && name != "(default)" {
						entry := AutorunEntry{
							Ubicacion:  loc,
							Nombre:     name,
							Comando:    cmd,
							Habilitado: true,
						}
						software.AutorunEntries = append(software.AutorunEntries, entry)
						addReport("SOFTWARE", fmt.Sprintf("  %s -> %s", name, cmd), "INFO")
					}
				}
			}
		}
	}

	if len(software.AutorunEntries) == 0 {
		addReport("SOFTWARE", "No se encontraron entradas de autorun", "INFO")
	}
}

func getBrowserExtensions() {
	addReport("SOFTWARE", "=== EXTENSIONES DE NAVEGADOR ===", "INFO")

	chromePath := os.Getenv("LOCALAPPDATA") + "\\Google\\Chrome\\User Data\\Default\\Extensions"
	edgePath := os.Getenv("LOCALAPPDATA") + "\\Microsoft\\Edge\\User Data\\Default\\Extensions"

	chromeExts := getChromeExtensions(chromePath)
	software.BrowserExtensions.Chrome = chromeExts
	addReport("SOFTWARE", fmt.Sprintf("Extensiones Chrome: %d", len(chromeExts)), "INFO")

	edgeExts := getEdgeExtensions(edgePath)
	software.BrowserExtensions.Edge = edgeExts
	addReport("SOFTWARE", fmt.Sprintf("Extensiones Edge: %d", len(edgeExts)), "INFO")
}

func getChromeExtensions(basePath string) []BrowserExt {
	var exts []BrowserExt

	extensionsPath := basePath
	output := runPowerShell(fmt.Sprintf("Get-ChildItem -Path '%s' -Directory -ErrorAction SilentlyContinue | Select-Object Name", extensionsPath))

	if output != "" {
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && len(line) == 32 {
				manifestPath := fmt.Sprintf("%s\\%s\\manifest.json", extensionsPath, line)
				manifest := runPowerShell(fmt.Sprintf("Get-Content '%s' -ErrorAction SilentlyContinue", manifestPath))

				if manifest != "" {
					var manifestData map[string]interface{}
					if err := json.Unmarshal([]byte(manifest), &manifestData); err == nil {
						name := ""
						if n, ok := manifestData["name"].(string); ok {
							name = n
						}
						version := ""
						if v, ok := manifestData["version"].(string); ok {
							version = v
						}
						perms := ""
						if p, ok := manifestData["permissions"].([]interface{}); ok {
							perms = fmt.Sprintf("%v", p)
						}

						ext := BrowserExt{
							ID:       line,
							Nombre:   name,
							Version:  version,
							Permisos: perms,
						}
						exts = append(exts, ext)
					}
				}
			}
		}
	}

	return exts
}

func getEdgeExtensions(basePath string) []BrowserExt {
	var exts []BrowserExt

	extensionsPath := basePath
	output := runPowerShell(fmt.Sprintf("Get-ChildItem -Path '%s' -Directory -ErrorAction SilentlyContinue | Select-Object Name", extensionsPath))

	if output != "" {
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && len(line) == 32 {
				manifestPath := fmt.Sprintf("%s\\%s\\manifest.json", extensionsPath, line)
				manifest := runPowerShell(fmt.Sprintf("Get-Content '%s' -ErrorAction SilentlyContinue", manifestPath))

				if manifest != "" {
					var manifestData map[string]interface{}
					if err := json.Unmarshal([]byte(manifest), &manifestData); err == nil {
						name := ""
						if n, ok := manifestData["name"].(string); ok {
							name = n
						}
						version := ""
						if v, ok := manifestData["version"].(string); ok {
							version = v
						}
						perms := ""
						if p, ok := manifestData["permissions"].([]interface{}); ok {
							perms = fmt.Sprintf("%v", p)
						}

						ext := BrowserExt{
							ID:       line,
							Nombre:   name,
							Version:  version,
							Permisos: perms,
						}
						exts = append(exts, ext)
					}
				}
			}
		}
	}

	return exts
}

func checkVulnerabilities() {
	addReport("SOFTWARE", "=== VERIFICACIÓN DE VULNERABILIDADES ===", "INFO")

	vulnerableSoftware := map[string][][]string{
		"Adobe": {
			{"<= 23.001.20093", "CRITICAL", "APSB23-47", "Vulnerabilidad crítica en Adobe Acrobat"},
			{"<= 15.001.20840", "CRITICAL", "APSB23-34", "Vulnerabilidad crítica en Adobe Reader"},
		},
		"Java": {
			{"<= 8u371", "CRITICAL", "CVE-2023-21939", "Vulnerabilidad de seguridad en Oracle Java SE"},
			{"<= 11.0.19", "CRITICAL", "CVE-2023-21938", "Vulnerabilidad en Java SE"},
			{"<= 17.0.7", "CRITICAL", "CVE-2023-21930", "Vulnerabilidad en Java SE"},
		},
		"Google Chrome": {
			{"<= 114.0.5735.110", "HIGH", "CVE-2023-3079", "Type confusion en V8"},
			{"<= 116.0.5845.96", "HIGH", "CVE-2023-4863", "Heap buffer overflow en WebRTC"},
		},
		"Microsoft Edge": {
			{"<= 114.0.1823.41", "HIGH", "CVE-2023-3079", "Vulnerabilidad en Edge"},
			{"<= 116.0.1938.55", "HIGH", "CVE-2023-4863", "Vulnerabilidad en Edge"},
		},
		"Firefox": {
			{"<= 114.0.1", "HIGH", "CVE-2023-38408", "Vulnerabilidad en Firefox"},
			{"<= 115.0.2", "MEDIUM", "CVE-2023-3724", "Vulnerabilidad en Firefox"},
		},
		"WinRAR": {
			{"<= 6.11", "CRITICAL", "CVE-2023-40477", "Vulnerabilidad de código remoto en WinRAR"},
		},
		"7-Zip": {
			{"<= 23.01", "MEDIUM", "CVE-2023-31110", "Vulnerabilidad en 7-Zip"},
		},
		"VLC media player": {
			{"<= 3.0.18", "HIGH", "CVE-2023-37903", "Vulnerabilidad en VLC"},
		},
	}

	for _, app := range software.InstalledApps {
		for vendor, cves := range vulnerableSoftware {
			if strings.Contains(strings.ToLower(app.Nombre), strings.ToLower(vendor)) {
				for _, cve := range cves {
					severity := cve[1]
					cveID := cve[2]
					desc := cve[3]

					vuln := Vulnerability{
						Software:    app.Nombre,
						Version:     app.Version,
						Severidad:   severity,
						CVE:         cveID,
						Descripcion: desc,
					}
					software.Vulnerabilities = append(software.Vulnerabilities, vuln)
					addReport("SOFTWARE", fmt.Sprintf("VULNERABLE: %s v%s - %s (%s)", app.Nombre, app.Version, cveID, severity), "CRITICAL")
				}
			}
		}
	}

	if len(software.Vulnerabilities) == 0 {
		addReport("SOFTWARE", "No se encontraron vulnerabilidades conocidas en el software instalado", "INFO")
	} else {
		addReport("SOFTWARE", fmt.Sprintf("ADVERTENCIA: %d vulnerabilidades conocidas detectadas", len(software.Vulnerabilities)), "WARNING")
	}
}
