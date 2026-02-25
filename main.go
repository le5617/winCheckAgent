package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
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
	Category  string `json:"categoria"`
	Level     string `json:"nivel"`
	Message   string `json:"mensaje"`
}

type DiagnosticReport struct {
	Fecha   string        `json:"fecha"`
	Report  []ReportEntry `json:"reporte"`
	Alertas []ReportEntry `json:"alertas"`
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

func addReport(category, message, level string) {
	entry := ReportEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Category:  category,
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
	addReport("SISTEMA", fmt.Sprintf("Sistema: %s %s", runtime.GOOS, runtime.GOARCH), "INFO")

	hostInfo, err := host.Info()
	if err == nil {
		addReport("SISTEMA", fmt.Sprintf("Hostname: %s", hostInfo.Hostname), "INFO")
		addReport("SISTEMA", fmt.Sprintf("Plataforma: %s %s", hostInfo.Platform, hostInfo.PlatformVersion), "INFO")
		addReport("SISTEMA", fmt.Sprintf("Kernel: %s", hostInfo.KernelVersion), "INFO")

		uptime := time.Duration(hostInfo.Uptime) * time.Second
		addReport("SISTEMA", fmt.Sprintf("Tiempo activo: %s", uptime.String()), "INFO")
	}

	vm, _ := mem.VirtualMemory()
	addReport("SISTEMA", fmt.Sprintf("RAM total: %.1f GB", float64(vm.Total)/(1024*1024*1024)), "INFO")
}

func checkCPU() {
	cpuPercent, _ := cpu.Percent(time.Second, false)
	avgPercent := 0.0
	if len(cpuPercent) > 0 {
		avgPercent = cpuPercent[0]
	}

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
		Fecha:   time.Now().Format("2006-01-02 15:04:05"),
		Report:  report,
		Alertas: alertas,
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
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("  REPORTE DE DIAGNOSTICO")
	fmt.Println(strings.Repeat("=", 60) + "\n")

	categorias := make(map[string][]ReportEntry)
	for _, entry := range report {
		categorias[entry.Category] = append(categorias[entry.Category], entry)
	}

	for cat, items := range categorias {
		fmt.Printf("\n--- %s ---\n", cat)
		for _, item := range items {
			emoji := map[string]string{
				"INFO":     "[+]",
				"WARNING":  "[!]",
				"CRITICAL": "[X]",
			}[item.Level]

			if emoji == "" {
				emoji = "[*]"
			}

			fmt.Printf("  %s %s\n", emoji, item.Message)
		}
	}

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

func runQuickDiagnostic() {
	fmt.Println("=" + strings.Repeat("=", 59))
	fmt.Println("  DIAGNOSTICO RAPIDO")
	fmt.Println("=" + strings.Repeat("=", 59))

	getSystemInfo()
	checkCPU()
	checkMemory()
	checkDisk()
	checkNetwork()
	checkBattery()
	checkProcesses()
	checkMalware()
	checkServices()
	generateRecommendations()

	printReport()

	filename := fmt.Sprintf("diagnostico_rapido_%s.json", time.Now().Format("20060102_150405"))
	saveReport(filename)
}

func runMonitor(duration int) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Printf("  MODO MONITOREO EN TIEMPO REAL (%d segundos)\n", duration)
	fmt.Println(strings.Repeat("=", 60) + "\n")

	for i := 0; i < duration; i++ {
		cpuPercent, _ := cpu.Percent(time.Second, false)
		vm, _ := mem.VirtualMemory()
		diskUsage, _ := disk.Usage("C:")

		level := "INFO"
		if len(cpuPercent) > 0 && (cpuPercent[0] > 90 || vm.UsedPercent > 90) {
			level = "CRITICAL"
		} else if len(cpuPercent) > 0 && (cpuPercent[0] > 80 || vm.UsedPercent > 80) {
			level = "WARNING"
		}

		cpuVal := 0.0
		if len(cpuPercent) > 0 {
			cpuVal = cpuPercent[0]
		}

		fmt.Printf("[%ds] CPU: %.1f%% | MEM: %.1f%% | DISCO: %.1f%%\n",
			i+1, cpuVal, vm.UsedPercent, diskUsage.UsedPercent)

		if level == "CRITICAL" {
			fmt.Println("  [!] ALERTA: Uso muy alto!")
		}

		time.Sleep(time.Second)
	}

	fmt.Printf("\n%s\n", strings.Repeat("=", 60))
	fmt.Println("  Fin del monitoreo")
	fmt.Printf("%s\n\n", strings.Repeat("=", 60))
}

func showAlerts() {
	if data, err := os.ReadFile("alertas_activas.json"); err == nil {
		var aa AlertasActivas
		if err := json.Unmarshal(data, &aa); err == nil {
			fmt.Println("\n=== ALERTAS ACTIVAS ===")
			for _, alerta := range aa.Alertas {
				fmt.Printf("[%s] %s\n", alerta.Level, alerta.Message)
			}
		}
	} else {
		fmt.Println("\nNo hay alertas guardadas")
	}
}

func showHistory() {
	if data, err := os.ReadFile("historial_diagnostico.json"); err == nil {
		var historial Historial
		if err := json.Unmarshal(data, &historial); err == nil {
			fmt.Println("\n=== HISTORIAL ===")
			// Show last 10 entries
			start := 0
			if len(historial.Historial) > 10 {
				start = len(historial.Historial) - 10
			}
			for i := start; i < len(historial.Historial); i++ {
				entry := historial.Historial[i]
				fmt.Printf("%s: CPU=%.1f%%, MEM=%.1f%%, CRITICAL=%d, WARNING=%d\n",
					entry.Timestamp[:10],
					entry.Resumen.CPUPromedio,
					entry.Resumen.MemoriaPorcentaje,
					entry.Resumen.AlertasCriticas,
					entry.Resumen.AlertasWarning)
			}
		}
	} else {
		fmt.Println("\nNo hay historial guardado")
	}
}

func showMenu() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("  CIBEROX ENDPOINT - MENU PRINCIPAL")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("\n1. Diagnostico completo")
	fmt.Println("2. Diagnostico rapido")
	fmt.Println("3. Auditoria de seguridad")
	fmt.Println("4. Monitoreo en tiempo real (10s)")
	fmt.Println("5. Ver alertas activas")
	fmt.Println("6. Ver historial")
	fmt.Println("7. Salir")

	fmt.Print("\nSelecciona una opcion: ")

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	option := scanner.Text()

	switch option {
	case "1":
		runFullDiagnostic()
	case "2":
		runQuickDiagnostic()
	case "3":
		report = nil
		alertas = nil
		runSecurityAudit()
		saveReport("auditoria_seguridad.json")
	case "4":
		runMonitor(10)
	case "5":
		showAlerts()
	case "6":
		showHistory()
	case "7":
		fmt.Println("\n¡Hasta luego!")
	default:
		fmt.Println("\nOpcion invalida")
	}
}

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--rapido", "-r":
			runQuickDiagnostic()
		case "--monitor", "-m":
			duration := 10
			if len(os.Args) > 2 {
				fmt.Sscanf(os.Args[2], "%d", &duration)
			}
			runMonitor(duration)
		case "--completo", "-c":
			runFullDiagnostic()
		default:
			fmt.Println("Uso: ciberox-endpoint [--rapido|--monitor [segundos]|--completo]")
		}
	} else {
		for {
			showMenu()
		}
	}
}

func getComputerInfo() {
	addReport("EQUIPO", "=== INFORMACIÓN DEL EQUIPO ===", "INFO")

	output := runPowerShell("(Get-CimInstance Win32_ComputerSystem).Name")
	if output != "" {
		addReport("EQUIPO", fmt.Sprintf("Nombre del equipo: %s", output), "INFO")
	}

	output = runPowerShell("(Get-CimInstance Win32_ComputerSystem).Domain")
	if output != "" {
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
			addReport("EQUIPO", fmt.Sprintf("Rol: %s", role), "INFO")
		}
	}

	output = runPowerShell("(Get-CimInstance Win32_OperatingSystem).Caption")
	if output != "" {
		addReport("EQUIPO", fmt.Sprintf("Sistema operativo: %s", output), "INFO")
	}

	output = runPowerShell("(Get-CimInstance Win32_OperatingSystem).Version")
	if output != "" {
		addReport("EQUIPO", fmt.Sprintf("Versión: %s", output), "INFO")
	}

	output = runPowerShell("(Get-CimInstance Win32_ComputerSystem).Manufacturer")
	if output != "" {
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
}
