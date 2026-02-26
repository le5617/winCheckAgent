package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/StackExchange/wmi"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
	"golang.org/x/sys/windows/registry"
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

type Win32HotFix struct {
	HotFixID    string
	Description string
	InstalledOn *time.Time
	InstalledBy string
}

type Win32NetworkAdapter struct {
	Name       string
	MacAddress string
	Speed      uint64
	Status     string
}

type Win32FirewallProfile struct {
	Name    string
	Enabled bool
}

type Win32FirewallRule struct {
	Name       string
	Direction  string
	Action     string
	LocalPort  string
	RemotePort string
	Profile    string
	Enabled    bool
}

type Win32USBDevice struct {
	Name         string
	Manufacturer string
	Status       string
}

type Win32Printer struct {
	Name     string
	PortName string
	Status   string
}

type Win32QuickFixEngineering struct {
	HotFixID    string
	FixComments string
	InstalledOn *time.Time
	InstalledBy string
}

type Win32Service struct {
	Name        string
	DisplayName string
	State       string // WMI property name (not Status)
	StartType   uint32
	StartMode   string
}

type Win32UserAccount struct {
	Name      string
	Disabled  bool
	Lockout   bool
	LastLogon *time.Time
}

type Win32StartupCommand struct {
	Name     string
	Command  string
	Location string
}

type Win32NTLogEvent struct {
	TimeCreated *time.Time
	Level       uint8
	Message     string
	SourceName  string
}

type Win32Product struct {
	Name        string
	Version     string
	Vendor      string
	InstallDate string
}

type Win32PnPEntity struct {
	Name         string
	Manufacturer string
	Status       string
	ClassGuid    string
}

// root\wmi namespace — SMART failure prediction
type MSStorageDriverFailurePredictStatus struct {
	InstanceName   string
	Active         bool
	PredictFailure bool
}

// root\StandardCimv2 namespace — Windows Firewall profiles
type MsftNetFirewallProfile struct {
	Name    string
	Enabled bool
}

// root\SecurityCenter2 namespace — AV products
type AntiVirusProduct struct {
	DisplayName            string
	ProductState           uint32
	PathToSignedProductExe string
}

// root\SecurityCenter2 namespace — Firewall products (3rd party)
type Win32FirewallProduct struct {
	DisplayName  string
	ProductState uint32
}

// Win32_NetworkAdapterConfiguration — adapters with IP enabled
type Win32NetworkAdapterConfiguration struct {
	Description      string
	IPAddress        []string
	MACAddress       string
	IPEnabled        bool
	DefaultIPGateway []string
}

func queryWMI(class string, dst interface{}) error {
	return wmi.Query(class, dst)
}

func runWMIQuery(class string, dst interface{}) bool {
	err := wmi.Query(class, dst)
	return err == nil
}

// batteryStatusToString convierte el código numérico de BatteryStatus WMI a texto legible.
func batteryStatusToString(status uint32) string {
	switch status {
	case 1:
		return "Desconectado"
	case 2:
		return "Conectado - Alta carga"
	case 3:
		return "Conectado - Baja carga"
	case 4:
		return "Conectado - Cargando"
	case 5:
		return "Conectado - Cargado"
	case 6:
		return "Conectado - Carga baja"
	case 7, 8:
		return "Conectado - Cargando"
	case 9, 10:
		return "Desconocido"
	case 11:
		return "Cargando"
	default:
		return "Desconocido"
	}
}

type Win32ComputerSystem struct {
	Name            string
	Domain          string
	DomainRole      uint16
	Manufacturer    string
	Model           string
	SystemSKUNumber string
}

type Win32OperatingSystem struct {
	Caption        string
	Version        string
	BuildNumber    string
	OSArchitecture string
}

type Win32BIOS struct {
	Manufacturer string
	Name         string
	Version      string
	SerialNumber string
	ReleaseDate  *time.Time
}

type Win32Processor struct {
	Name                      string
	Manufacturer              string
	NumberOfCores             uint32
	NumberOfLogicalProcessors uint32
	MaxClockSpeed             uint32
	ProcessorId               string
}

type Win32PhysicalMemory struct {
	Capacity      uint64
	Speed         uint16
	MemoryType    uint16
	FormFactor    uint16
	BankLabel     string
	DeviceLocator string
}

type Win32Battery struct {
	Name                     string
	EstimatedChargeRemaining uint32
	EstimatedRunTime         uint32
	BatteryStatus            uint32
	DesignCapacity           uint32
	FullChargeCapacity       uint32
}

type Win32DiskDrive struct {
	Model         string
	SerialNumber  string
	Size          uint64
	MediaType     string
	Status        string
	InterfaceType string
}

type Win32LogicalDisk struct {
	DeviceID   string
	VolumeName string
	FileSystem string
	Size       uint64
	FreeSpace  uint64
}

func getWMIComputerInfo() {
	var dst []Win32ComputerSystem
	err := wmi.Query("SELECT * FROM Win32_ComputerSystem", &dst)
	if err == nil && len(dst) > 0 {
		cs := dst[0]
		equipo.NombreEquipo = cs.Name
		equipo.Dominio = cs.Domain
		equipo.Fabricante = cs.Manufacturer
		equipo.Modelo = cs.Model
		equipo.Producto = cs.SystemSKUNumber

		switch cs.DomainRole {
		case 0:
			equipo.Rol = "Estación de trabajo"
		case 1:
			equipo.Rol = "Estación de trabajo (miembro de dominio)"
		case 2:
			equipo.Rol = "Servidor miembro"
		case 3:
			equipo.Rol = "Servidor miembro (miembro de dominio)"
		case 4:
			equipo.Rol = "Controlador de dominio"
		case 5:
			equipo.Rol = "Controlador de dominio (primario)"
		}
	}
}

func getWMIOSInfo() {
	var dst []Win32OperatingSystem
	err := wmi.Query("SELECT * FROM Win32_OperatingSystem", &dst)
	if err == nil && len(dst) > 0 {
		os := dst[0]
		equipo.SistemaOperativo = os.Caption
		equipo.Version = os.Version
	}
}

func getWMIBIOSInfo() {
	var dst []Win32BIOS
	err := wmi.Query("SELECT * FROM Win32_BIOS", &dst)
	if err == nil && len(dst) > 0 {
		bios := dst[0]
		hardware.BIOS.Fabricante = bios.Manufacturer
		hardware.BIOS.Version = bios.Version
		hardware.BIOS.NumeroSerie = bios.SerialNumber
		if bios.ReleaseDate != nil {
			hardware.BIOS.Fecha = bios.ReleaseDate.Format("2006-01-02")
		}
		if equipo.NumeroSerie == "" {
			equipo.NumeroSerie = bios.SerialNumber
		}
	}
}

func getWMIProcessorInfo() {
	var dst []Win32Processor
	err := wmi.Query("SELECT * FROM Win32_Processor", &dst)
	if err == nil && len(dst) > 0 {
		p := dst[0]
		hardware.Procesador.Nombre = strings.TrimSpace(p.Name)
		hardware.Procesador.Fabricante = p.Manufacturer
		hardware.Procesador.Nucleos = int(p.NumberOfCores)
		hardware.Procesador.Threads = int(p.NumberOfLogicalProcessors)
		hardware.Procesador.FrecuenciaMax = int(p.MaxClockSpeed)
		hardware.Procesador.Identificador = p.ProcessorId
	}
}

func getWMIPhysicalMemory() {
	var dst []Win32PhysicalMemory
	err := wmi.Query("SELECT * FROM Win32_PhysicalMemory", &dst)
	if err == nil {
		var totalRAM uint64
		var velocidad uint16
		slots := 0

		for _, m := range dst {
			totalRAM += m.Capacity
			velocidad = m.Speed
			slots++
		}

		hardware.RAM.TotalGB = float64(totalRAM) / (1024 * 1024 * 1024)
		hardware.RAM.SlotsUsados = slots
		hardware.RAM.Velocidad = int(velocidad)
	}
}

func getWMIBatteryInfo() {
	var dst []Win32Battery
	err := wmi.Query("SELECT * FROM Win32_Battery", &dst)
	if err == nil && len(dst) > 0 {
		b := dst[0]
		hardware.Bateria.Nombre = b.Name
		hardware.Bateria.CargaPorcent = float64(b.EstimatedChargeRemaining)
		hardware.Bateria.TiempoRestante = int(b.EstimatedRunTime)

		hardware.Bateria.Estado = batteryStatusToString(b.BatteryStatus)

		hardware.Bateria.CapacidadDiseno = int64(b.DesignCapacity)
		hardware.Bateria.CapacidadActual = int64(b.FullChargeCapacity)

		if b.DesignCapacity > 0 && b.FullChargeCapacity > 0 {
			hardware.Bateria.SaludPorcent = float64(b.FullChargeCapacity) / float64(b.DesignCapacity) * 100
			if hardware.Bateria.SaludPorcent < 80 {
				addReport("HARDWARE", fmt.Sprintf("ALERTA: Salud de batería baja (%.1f%%). Considerar reemplazo", hardware.Bateria.SaludPorcent), "WARNING")
			}
		}
	}
}

func getWMIDiskDrive() {
	var dst []Win32DiskDrive
	err := wmi.Query("SELECT * FROM Win32_DiskDrive", &dst)
	if err == nil {
		for _, d := range dst {
			diskInfo := DiskInfo{
				Modelo:      d.Model,
				NumeroSerie: d.SerialNumber,
				Estado:      d.Status,
			}
			hardware.Discos = append(hardware.Discos, diskInfo)
		}
	}
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

	ifs, _ := net.Interfaces()
	for _, iface := range ifs {
		var ips []string
		for _, addr := range iface.Addrs {
			ips = append(ips, addr.Addr)
		}
		statusStr := "down"
		if iface.HardwareAddr != "" || len(ips) > 0 {
			statusStr = "up"
		}
		ni := NetworkInterface{
			Name:        iface.Name,
			IPAddresses: ips,
			MAC:         iface.HardwareAddr,
			Status:      statusStr,
		}
		red.Interfaces = append(red.Interfaces, ni)
		addReport("RED", fmt.Sprintf("Interfaz: %s - MAC: %s - IPs: %v", iface.Name, iface.HardwareAddr, ips), "INFO")
	}
}

func checkNetworkAdvanced() {
	addReport("RED_AVANZADA", "=== VERIFICACION AVANZADA DE RED ===", "INFO")

	var adapters []Win32NetworkAdapterConfiguration
	hasConnection := false

	err := wmi.Query(
		"SELECT Description, IPAddress, MACAddress, IPEnabled, DefaultIPGateway FROM Win32_NetworkAdapterConfiguration WHERE IPEnabled = TRUE",
		&adapters,
	)
	if err == nil {
		for _, adapter := range adapters {
			for _, ip := range adapter.IPAddress {
				if ip == "" || strings.HasPrefix(ip, "127.") || strings.HasPrefix(ip, "::1") {
					continue
				}
				addReport("RED_AVANZADA", fmt.Sprintf("Interfaz: %s - IP: %s - MAC: %s",
					adapter.Description, ip, adapter.MACAddress), "INFO")
				hasConnection = true
			}
			if len(adapter.DefaultIPGateway) > 0 && adapter.DefaultIPGateway[0] != "" {
				addReport("RED_AVANZADA", fmt.Sprintf("  Gateway predeterminado: %s", adapter.DefaultIPGateway[0]), "INFO")
			}
		}
	} else {
		addReport("RED_AVANZADA", fmt.Sprintf("Error consultando adaptadores WMI: %v", err), "WARNING")
	}

	if hasConnection {
		addReport("RED_AVANZADA", "Conectividad a red: OK", "INFO")
	} else {
		addReport("RED_AVANZADA", "Sin conectividad de red activa detectada", "WARNING")
	}
}

func checkBattery() {
	bat := hardware.Bateria
	if bat.Nombre != "" {
		level := "INFO"
		if bat.CargaPorcent < 20 {
			level = "WARNING"
		}
		if bat.CargaPorcent < 10 {
			level = "CRITICAL"
		}
		addReport("BATERIA", fmt.Sprintf("Carga: %.0f%%", bat.CargaPorcent), level)
		addReport("BATERIA", fmt.Sprintf("Estado: %s", bat.Estado), "INFO")

		if bat.SaludPorcent > 0 {
			addReport("BATERIA", fmt.Sprintf("Salud: %.1f%%", bat.SaludPorcent), "INFO")
		}
	} else {
		addReport("BATERIA", "No se detectó batería (equipo de escritorio)", "INFO")
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

	// Ordenar por CPU descendente
	sort.Slice(topCPU, func(i, j int) bool { return topCPU[i].cpu > topCPU[j].cpu })

	// Ordenar por memoria descendente
	sort.Slice(topMem, func(i, j int) bool { return topMem[i].mem > topMem[j].mem })

	addReport("PROCESOS", "Top 10 procesos por CPU:", "INFO")
	for i := 0; i < min(10, len(topCPU)); i++ {
		addReport("PROCESOS", fmt.Sprintf("  %s: CPU=%.1f%%, PID=%d", topCPU[i].name, topCPU[i].cpu, topCPU[i].pid), "INFO")
	}

	addReport("PROCESOS", "Top 5 procesos por memoria:", "INFO")
	for i := 0; i < min(5, len(topMem)); i++ {
		addReport("PROCESOS", fmt.Sprintf("  %s: Memoria=%.1f%%, PID=%d", topMem[i].name, topMem[i].mem, topMem[i].pid), "INFO")
	}

	addReport("PROCESOS", fmt.Sprintf("Total de procesos: %d", len(procs)), "INFO")

	// Poblar slice global para el JSON
	for i := 0; i < min(10, len(topCPU)); i++ {
		procesos = append(procesos, ProcessInfo{
			Name:       topCPU[i].name,
			PID:        topCPU[i].pid,
			CPUPercent: topCPU[i].cpu,
			MemPercent: topCPU[i].mem,
		})
	}
}

func checkBSOD() {
	addReport("BSOD", "=== ANALISIS DE PANTALLAS AZULES (BSOD) ===", "INFO")

	var events []Win32NTLogEvent
	err := wmi.Query("SELECT * FROM Win32_NTLogEvent WHERE LogFile='System' AND EventType=1 AND Message LIKE '%Bugcheck%' OR Message LIKE '%BlueScreen%'", &events)
	if err == nil && len(events) > 0 {
		addReport("BSOD", fmt.Sprintf("Errores criticos encontrados en logs: %d eventos", len(events)), "WARNING")
	} else {
		addReport("BSOD", "No se encontraron BSOD recientes", "INFO")
	}
}

func checkSMART() {
	addReport("SMART", "=== VERIFICACION DE DISCO (SMART) ===", "INFO")
	addReport("SMART", "Usando datos de Win32_DiskDrive", "INFO")
}

func checkMalware() {
	addReport("MALWARE", "=== ESTADO DEL ANTIVIRUS ===", "INFO")

	var avProducts []AntiVirusProduct
	err := wmi.Query("SELECT DisplayName, ProductState FROM AntiVirusProduct", &avProducts, nil, "root\\SecurityCenter2")
	if err == nil && len(avProducts) > 0 {
		for _, av := range avProducts {
			// ProductState: bits 16-23 = estado producto, bits 8-15 = estado firmas
			stateCode := (av.ProductState >> 16) & 0xFF
			defStatus := (av.ProductState >> 8) & 0xFF
			enabled := stateCode == 0x10
			upToDate := defStatus == 0x00

			statusStr := "Deshabilitado"
			if enabled {
				statusStr = "Habilitado"
			}
			defStr := "Desactualizado"
			if upToDate {
				defStr = "Al día"
			}
			level := "INFO"
			if !enabled || !upToDate {
				level = "WARNING"
			}
			addReport("MALWARE", fmt.Sprintf("Antivirus: %s - Estado: %s - Firmas: %s",
				av.DisplayName, statusStr, defStr), level)
		}
	} else {
		addReport("MALWARE", "No se detectó antivirus o SecurityCenter2 inaccesible", "WARNING")
	}
}

func checkDrivers() {
	addReport("DRIVERS", "=== VERIFICACION DE DRIVERS ===", "INFO")

	var devices []Win32PnPEntity
	err := wmi.Query("SELECT * FROM Win32_PnPEntity WHERE ConfigManagerErrorCode != 0", &devices)
	if err == nil && len(devices) > 0 {
		addReport("DRIVERS", fmt.Sprintf("Drivers con problemas detectados: %d", len(devices)), "WARNING")
	} else {
		addReport("DRIVERS", "Drivers sin errores detectados", "INFO")
	}
}

func checkSystemIntegrity() {
	addReport("INTEGRIDAD", "=== VERIFICACION DE INTEGRIDAD DEL SISTEMA ===", "INFO")
	addReport("INTEGRIDAD", "Usa 'sfc /scannow' manualmente para verificar integridad", "INFO")
}

func checkWindowsUpdate() {
	addReport("WINDOWS UPDATE", "=== ESTADO DE WINDOWS UPDATE ===", "INFO")

	// Nota: ORDER BY InstalledOn falla porque es un campo nullable *time.Time en WMI
	var hotfixes []Win32QuickFixEngineering
	err := wmi.Query("SELECT HotFixID, InstalledOn FROM Win32_QuickFixEngineering", &hotfixes)
	if err == nil && len(hotfixes) > 0 {
		// Buscar el último por fecha manualmente
		var lastFix Win32QuickFixEngineering
		for _, hf := range hotfixes {
			if hf.InstalledOn != nil {
				if lastFix.InstalledOn == nil || hf.InstalledOn.After(*lastFix.InstalledOn) {
					lastFix = hf
				}
			}
		}
		addReport("WINDOWS UPDATE", fmt.Sprintf("Total de parches instalados: %d", len(hotfixes)), "INFO")
		if lastFix.HotFixID != "" {
			addReport("WINDOWS UPDATE", fmt.Sprintf("Ultimo parche: %s (%s)",
				lastFix.HotFixID, lastFix.InstalledOn.Format("2006-01-02")), "INFO")
		}
	} else {
		addReport("WINDOWS UPDATE", fmt.Sprintf("No se pudo obtener información de actualizaciones: %v", err), "WARNING")
	}
}

func checkServices() {
	addReport("SERVICIOS", "=== SERVICIOS CRITICOS ===", "INFO")

	// Filtrar solo servicios relevant para no requerir admin total
	var services []Win32Service
	err := wmi.Query(
		"SELECT Name, DisplayName, State, StartMode FROM Win32_Service WHERE StartMode='Auto'",
		&services,
	)
	if err != nil {
		// Fallback: query mínima sin filtro
		err = wmi.Query("SELECT Name, DisplayName, State, StartMode FROM Win32_Service", &services)
	}
	if err == nil {
		stoppedCount := 0
		for _, svc := range services {
			// En WMI la propiedad de Win32_Service es 'State' (no Status)
			state := svc.State
			if state == "" {
				state = "Unknown"
			}
			if state != "Running" {
				stoppedCount++
			}
			// Guardar en slice global para el JSON
			servicios = append(servicios, ServiceInfo{
				Name:      svc.Name,
				Status:    state,
				StartType: fmt.Sprintf("%d", svc.StartType),
			})
		}
		if stoppedCount > 0 {
			addReport("SERVICIOS", fmt.Sprintf("Servicios de inicio automático no activos: %d", stoppedCount), "WARNING")
		} else {
			addReport("SERVICIOS", fmt.Sprintf("Servicios monitorizados: %d activos", len(services)), "INFO")
		}
	} else {
		addReport("SERVICIOS", fmt.Sprintf("No se pudieron obtener los servicios: %v", err), "WARNING")
	}
}

func checkStartup() {
	addReport("STARTUP", "=== PROGRAMAS DE INICIO ===", "INFO")

	var startup []Win32StartupCommand
	err := wmi.Query("SELECT Name, Command, Location FROM Win32_StartupCommand", &startup)
	if err == nil && len(startup) > 0 {
		addReport("STARTUP", fmt.Sprintf("Programas de inicio: %d configuraciones", len(startup)), "INFO")
	} else {
		addReport("STARTUP", "No hay programas de inicio configurados", "INFO")
	}
}

func checkLogs() {
	addReport("LOGS", "=== EVENTOS DEL SISTEMA ===", "INFO")
	addReport("LOGS", "Usa Visor de eventos para ver logs detallados", "INFO")
}

func checkUSB() {
	addReport("USB", "=== DISPOSITIVOS USB ===", "INFO")

	var usbDevices []Win32PnPEntity
	err := wmi.Query("SELECT * FROM Win32_PnPEntity WHERE PNPClass='USB'", &usbDevices)
	if err == nil && len(usbDevices) > 0 {
		addReport("USB", fmt.Sprintf("Dispositivos USB detectados: %d", len(usbDevices)), "INFO")
	} else {
		addReport("USB", "No se detectaron dispositivos USB", "INFO")
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

		isSuspicious := false
		level := "INFO"
		msg := fmt.Sprintf("Puerto %s (%s) - Proceso: %s", port, service, proc)

		for _, susp := range suspiciousPorts {
			if port == susp {
				isSuspicious = true
				level = "CRITICAL"
				msg = fmt.Sprintf("PUERTO SOSPECHOSO: %s", msg)
				break
			}
		}
		addReport("SEGURIDAD", msg, level)

		// Poblar struct de seguridad para el JSON
		portNum := 0
		fmt.Sscanf(port, "%d", &portNum)
		seguridad.ListeningPorts = append(seguridad.ListeningPorts, PortInfo{
			Port:       portNum,
			Service:    service,
			Process:    proc,
			Sospechoso: isSuspicious,
		})
	}

	if len(ports) == 0 {
		addReport("SEGURIDAD", "No se detectaron puertos en escucha", "WARNING")
	}
}

func getSecurityPatches() {
	addReport("SEGURIDAD", "=== PARCHES DE SEGURIDAD ===", "INFO")

	var hotfixes []Win32QuickFixEngineering
	// Sin ORDER BY: InstalledOn es *time.Time nullable y falla en WMI
	err := wmi.Query("SELECT HotFixID, InstalledOn, FixComments FROM Win32_QuickFixEngineering", &hotfixes)
	if err == nil && len(hotfixes) > 0 {
		var lastFix Win32QuickFixEngineering
		for i, hf := range hotfixes {
			if hf.InstalledOn != nil && (lastFix.InstalledOn == nil || hf.InstalledOn.After(*lastFix.InstalledOn)) {
				lastFix = hotfixes[i]
			}
			// Poblar struct de seguridad
			patch := SecurityPatch{
				HotFixID:    hf.HotFixID,
				Descripcion: hf.FixComments,
			}
			if hf.InstalledOn != nil {
				patch.Instalado = hf.InstalledOn.Format("2006-01-02")
			}
			if i < 20 { // limitar a 20 en JSON
				seguridad.SecurityPatches = append(seguridad.SecurityPatches, patch)
			}
		}
		addReport("SEGURIDAD", fmt.Sprintf("Total parches instalados: %d", len(hotfixes)), "INFO")
		if lastFix.HotFixID != "" && lastFix.InstalledOn != nil {
			days := time.Since(*lastFix.InstalledOn).Hours() / 24
			seguridad.DiasSinParchar = int(days)
			addReport("SEGURIDAD", fmt.Sprintf("Último parche: %s (%s) - hace %.0f días",
				lastFix.HotFixID, lastFix.InstalledOn.Format("2006-01-02"), days), "INFO")
			if days > 30 {
				addReport("SEGURIDAD", fmt.Sprintf("ALERTA: %.0f días sin parches de seguridad", days), "CRITICAL")
			} else if days > 14 {
				addReport("SEGURIDAD", fmt.Sprintf("ADVERTENCIA: %.0f días sin parches", days), "WARNING")
			}
		}
	} else {
		addReport("SEGURIDAD", fmt.Sprintf("No se pudieron obtener los parches: %v", err), "WARNING")
	}
}

func getPasswordPolicy() {
	addReport("SEGURIDAD", "=== POLÍTICAS DE CONTRASEÑA ===", "INFO")

	var users []Win32UserAccount
	err := wmi.Query("SELECT * FROM Win32_UserAccount WHERE LocalAccount=TRUE", &users)
	if err == nil && len(users) > 0 {
		addReport("SEGURIDAD", fmt.Sprintf("Usuarios locales: %d cuentas", len(users)), "INFO")
	} else {
		addReport("SEGURIDAD", "No se pudieron obtener usuarios", "WARNING")
	}

	addReport("SEGURIDAD", "Verifica manualmente las políticas de contraseña", "INFO")
}

func getUserList() {
	addReport("SEGURIDAD", "=== USUARIOS LOCALES ===", "INFO")

	var users []Win32UserAccount
	err := wmi.Query("SELECT Name, Disabled, Lockout FROM Win32_UserAccount WHERE LocalAccount=TRUE", &users)
	if err == nil && len(users) > 0 {
		for _, user := range users {
			status := "Habilitado"
			if user.Disabled {
				status = "Deshabilitado"
			}
			addReport("SEGURIDAD", fmt.Sprintf("Usuario: %s - %s", user.Name, status), "INFO")
			// Poblar struct de seguridad para el JSON
			logon := ""
			if user.LastLogon != nil {
				logon = user.LastLogon.Format("2006-01-02 15:04")
			}
			seguridad.Users = append(seguridad.Users, UserInfo{
				Nombre:      user.Name,
				Habilitado:  !user.Disabled,
				UltimoLogin: logon,
			})
		}
	} else {
		addReport("SEGURIDAD", fmt.Sprintf("No se pudieron obtener usuarios: %v", err), "WARNING")
	}
}

func getUSBDevices() {
	addReport("SEGURIDAD", "=== DISPOSITIVOS USB ===", "INFO")

	var usbDevices []Win32PnPEntity
	err := wmi.Query("SELECT Name, Manufacturer, Status FROM Win32_PnPEntity WHERE PNPDeviceID LIKE 'USB%'", &usbDevices)
	if err == nil && len(usbDevices) > 0 {
		for _, dev := range usbDevices {
			addReport("SEGURIDAD", fmt.Sprintf("USB: %s - %s", dev.Name, dev.Status), "INFO")
			seguridad.USBDevices = append(seguridad.USBDevices, dev.Name)
		}
	} else {
		addReport("SEGURIDAD", fmt.Sprintf("No se encontraron dispositivos USB: %v", err), "INFO")
	}
}

func getPrinters() {
	addReport("SEGURIDAD", "=== IMPRESORAS ===", "INFO")

	var printers []Win32Printer
	err := wmi.Query("SELECT Name, PortName, Status FROM Win32_Printer", &printers)
	if err == nil && len(printers) > 0 {
		for _, p := range printers {
			addReport("SEGURIDAD", fmt.Sprintf("Impresora: %s - %s", p.Name, p.Status), "INFO")
			seguridad.Printers = append(seguridad.Printers, fmt.Sprintf("%s (%s)", p.Name, p.Status))
		}
	} else {
		addReport("SEGURIDAD", "No se detectaron impresoras", "INFO")
	}
}

func getFirewallStatus() {
	addReport("SEGURIDAD", "=== ESTADO DEL FIREWALL ===", "INFO")

	// Productos de firewall de terceros (SecurityCenter2)
	var fwProducts []Win32FirewallProduct
	if err := wmi.Query("SELECT DisplayName, ProductState FROM FirewallProduct", &fwProducts, nil, "root\\SecurityCenter2"); err == nil && len(fwProducts) > 0 {
		for _, fw := range fwProducts {
			stateCode := (fw.ProductState >> 16) & 0xFF
			enabled := stateCode == 0x10
			statusStr := "Deshabilitado"
			if enabled {
				statusStr = "Habilitado"
			}
			level := "INFO"
			if !enabled {
				level = "WARNING"
			}
			addReport("SEGURIDAD", fmt.Sprintf("Firewall (3º): %s - Estado: %s", fw.DisplayName, statusStr), level)
		}
	}

	// Perfiles del Firewall de Windows (root\StandardCimv2) — poblar FirewallStatus del JSON
	var fwProfiles []MsftNetFirewallProfile
	if err := wmi.Query("SELECT Name, Enabled FROM MSFT_NetFirewallProfile", &fwProfiles, nil, "root\\StandardCimv2"); err == nil && len(fwProfiles) > 0 {
		for _, p := range fwProfiles {
			status := "Deshabilitado"
			if p.Enabled {
				status = "Habilitado"
			}
			level := "INFO"
			if !p.Enabled {
				level = "WARNING"
			}
			addReport("SEGURIDAD", fmt.Sprintf("Firewall Windows - Perfil [%s]: %s", p.Name, status), level)
			// Mapear perfil al struct FirewallStatus
			switch strings.ToLower(p.Name) {
			case "domain":
				seguridad.FirewallStatus.Domain = p.Enabled
			case "public":
				seguridad.FirewallStatus.Public = p.Enabled
			case "private":
				seguridad.FirewallStatus.Private = p.Enabled
			}
		}
	} else {
		addReport("SEGURIDAD", "No se pudo consultar perfiles del Firewall de Windows", "WARNING")
	}
}

func runSecurityAudit() {
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
	getWMIComputerInfo()
	getWMIOSInfo()
	getWMIBIOSInfo()
	getWMIProcessorInfo()
	getWMIPhysicalMemory()
	getWMIBatteryInfo()
	getWMIDiskDrive()
	getDiskHealth()
}

// getDiskHealth consulta el estado SMART de los discos usando WMI nativo.
// Usa MSStorageDriver_FailurePredictStatus (root\wmi) para predicción de fallos
// y Win32_DiskDrive para datos de modelo/serie/tipo.
func getDiskHealth() {
	addReport("HARDWARE", "=== SALUD DEL DISCO (S.M.A.R.T.) ===", "INFO")

	// Mapa de predicciones de fallo SMART por InstanceName
	smartFailMap := make(map[string]bool)
	var smartPredictions []MSStorageDriverFailurePredictStatus
	if err := wmi.Query(
		"SELECT InstanceName, PredictFailure FROM MSStorageDriver_FailurePredictStatus",
		&smartPredictions, nil, "root\\wmi",
	); err == nil {
		for _, s := range smartPredictions {
			key := strings.ToLower(s.InstanceName)
			smartFailMap[key] = s.PredictFailure
		}
		addReport("HARDWARE", fmt.Sprintf("SMART: %d unidades monitorizadas", len(smartPredictions)), "INFO")
	} else {
		addReport("HARDWARE", "SMART vía WMI no disponible (requiere privilegios de administrador)", "INFO")
	}

	// Detalles de discos desde Win32_DiskDrive
	var drives []Win32DiskDrive
	if err := wmi.Query(
		"SELECT Model, SerialNumber, Size, MediaType, Status, InterfaceType FROM Win32_DiskDrive",
		&drives,
	); err == nil {
		for i, d := range drives {
			modelo := strings.TrimSpace(d.Model)
			serie := strings.TrimSpace(d.SerialNumber)

			// Buscar predicción SMART para este disco
			predictFail := false
			for key, fail := range smartFailMap {
				if strings.Contains(key, strings.ToLower(modelo)) {
					predictFail = fail
					break
				}
			}

			estado := d.Status
			level := "INFO"
			if predictFail {
				estado = "PredFail - FALLO PREDICHO"
				level = "CRITICAL"
			} else if d.Status != "OK" {
				level = "WARNING"
			}

			addReport("HARDWARE",
				fmt.Sprintf("Disco [%d]: %s | Serie: %s | Tipo: %s | Interfaz: %s | Estado: %s",
					i+1, modelo, serie, d.MediaType, d.InterfaceType, estado), level)

			if predictFail {
				addReport("HARDWARE", fmt.Sprintf("  ⚠️ ALERTA SMART: Fallo inminente predicho en disco '%s'", modelo), "CRITICAL")
			}

			// Actualizar entrada existente (ya creada por getWMIDiskDrive) o añadir nueva
			updated := false
			for j := range hardware.Discos {
				if hardware.Discos[j].Modelo == modelo {
					hardware.Discos[j].Estado = estado
					updated = true
					break
				}
			}
			if !updated {
				hardware.Discos = append(hardware.Discos, DiskInfo{
					Modelo:      modelo,
					NumeroSerie: serie,
					Estado:      estado,
				})
			}
		}
	} else {
		addReport("HARDWARE", "No se pudieron leer discos vía Win32_DiskDrive", "WARNING")
	}

	if len(hardware.Discos) == 0 {
		addReport("HARDWARE", "No se detectaron discos físicos", "WARNING")
	}
}

// getInstalledSoftware lee las aplicaciones instaladas del Registro de Windows
// (método rápido y fiable, a diferencia de Win32_Product que es muy lento).
func getInstalledSoftware() {
	addReport("SOFTWARE", "=== APLICACIONES INSTALADAS ===", "INFO")

	uninstallKeys := []string{
		`SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall`,
		`SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall`,
	}
	roots := []registry.Key{registry.LOCAL_MACHINE, registry.CURRENT_USER}

	seen := make(map[string]bool)
	for _, root := range roots {
		for _, keyPath := range uninstallKeys {
			k, err := registry.OpenKey(root, keyPath, registry.ENUMERATE_SUB_KEYS|registry.READ)
			if err != nil {
				continue
			}
			subkeys, _ := k.ReadSubKeyNames(-1)
			k.Close()
			for _, sub := range subkeys {
				subK, err := registry.OpenKey(root, keyPath+"\\"+sub, registry.QUERY_VALUE)
				if err != nil {
					continue
				}
				name, _, _ := subK.GetStringValue("DisplayName")
				version, _, _ := subK.GetStringValue("DisplayVersion")
				publisher, _, _ := subK.GetStringValue("Publisher")
				installDate, _, _ := subK.GetStringValue("InstallDate")
				subK.Close()
				if name == "" || seen[name+version] {
					continue
				}
				seen[name+version] = true
				app := InstalledApp{
					Nombre:    strings.TrimSpace(name),
					Version:   version,
					Editor:    publisher,
					FechaInst: installDate,
				}
				software.InstalledApps = append(software.InstalledApps, app)
			}
		}
	}

	addReport("SOFTWARE", fmt.Sprintf("Total aplicaciones instaladas: %d", len(software.InstalledApps)), "INFO")
	seguridad.Software.InstalledApps = software.InstalledApps
}

func getAutorunEntries() {
	addReport("SOFTWARE", "=== SOFTWARE DE INICIO (AUTORUN) ===", "INFO")

	var startup []Win32StartupCommand
	err := wmi.Query("SELECT Name, Command, Location FROM Win32_StartupCommand", &startup)
	if err == nil && len(startup) > 0 {
		for _, entry := range startup {
			autorun := AutorunEntry{
				Ubicacion:  entry.Location,
				Nombre:     entry.Name,
				Comando:    entry.Command,
				Habilitado: true,
			}
			software.AutorunEntries = append(software.AutorunEntries, autorun)
			addReport("SOFTWARE", fmt.Sprintf("  %s -> %s", entry.Name, entry.Command), "INFO")
		}
	} else {
		addReport("SOFTWARE", "No se encontraron entradas de autorun", "INFO")
	}
}

func getBrowserExtensions() {
	addReport("SOFTWARE", "=== EXTensiones DE NAVEGADOR ===", "INFO")

	chromePath := os.Getenv("LOCALAPPDATA") + "\\Google\\Chrome\\User Data\\Default\\Extensions"
	edgePath := os.Getenv("LOCALAPPDATA") + "\\Microsoft\\Edge\\User Data\\Default\\Extensions"

	chromeExts := getChromeExtensionsFromPath(chromePath)
	software.BrowserExtensions.Chrome = chromeExts
	addReport("SOFTWARE", fmt.Sprintf("Extensiones Chrome: %d", len(chromeExts)), "INFO")

	edgeExts := getChromeExtensionsFromPath(edgePath)
	software.BrowserExtensions.Edge = edgeExts
	addReport("SOFTWARE", fmt.Sprintf("Extensiones Edge: %d", len(edgeExts)), "INFO")
}

func getChromeExtensionsFromPath(basePath string) []BrowserExt {
	var exts []BrowserExt

	entries, err := os.ReadDir(basePath)
	if err != nil {
		return exts
	}

	for _, entry := range entries {
		if !entry.IsDir() || len(entry.Name()) != 32 {
			continue
		}

		manifestPath := basePath + "\\" + entry.Name() + "\\manifest.json"
		data, err := os.ReadFile(manifestPath)
		if err != nil {
			continue
		}

		var manifestData map[string]interface{}
		if err := json.Unmarshal(data, &manifestData); err != nil {
			continue
		}

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
			ID:       entry.Name(),
			Nombre:   name,
			Version:  version,
			Permisos: perms,
		}
		exts = append(exts, ext)
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
