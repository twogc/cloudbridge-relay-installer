/*
YARA Rules для CloudBridge Relay Security
Детекция вредоносного ПО и подозрительных файлов
*/

// Правило для детекции подозрительных скриптов
rule Suspicious_Scripts {
    meta:
        description = "Обнаружение подозрительных скриптов"
        author = "CloudBridge Security Team"
        date = "2024-01-01"
        severity = "medium"
        tags = "malware, script, execution"
    
    strings:
        $shell_commands = /(bash|sh|python|perl|ruby|php)\s+(-c|-e)/
        $suspicious_cmds = /(wget|curl|nc|netcat|telnet|ssh|scp)\s+/
        $encoded_strings = /base64\s+-d/
        $eval_commands = /eval\s*\(/
        $exec_commands = /exec\s*\(/
        
    condition:
        any of them
}

// Правило для детекции криптомайнеров
rule Cryptominer_Detection {
    meta:
        description = "Обнаружение криптомайнеров"
        author = "CloudBridge Security Team"
        date = "2024-01-01"
        severity = "high"
        tags = "malware, cryptominer, mining"
    
    strings:
        $mining_pools = /(stratum|pool|mine|mining)/
        $crypto_algorithms = /(sha256|scrypt|x11|ethash|randomx)/
        $mining_software = /(xmrig|xmr-stak|cpuminer|ccminer)/
        $gpu_commands = /(nvidia-smi|amd-smi|gpu)/
        
    condition:
        2 of them
}

// Правило для детекции бэкдоров
rule Backdoor_Detection {
    meta:
        description = "Обнаружение бэкдоров"
        author = "CloudBridge Security Team"
        date = "2024-01-01"
        severity = "critical"
        tags = "malware, backdoor, persistence"
    
    strings:
        $bind_shell = /bind\s*\(.*\s*,\s*.*\s*,\s*.*\s*\)/
        $reverse_shell = /(nc|netcat).*?-e\s+bash/
        $suspicious_ports = /(4444|8080|1337|31337)/
        $shell_commands = /(/bin/bash|/bin/sh|cmd.exe)/
        
    condition:
        2 of them
}

// Правило для детекции кейлоггеров
rule Keylogger_Detection {
    meta:
        description = "Обнаружение кейлоггеров"
        author = "CloudBridge Security Team"
        date = "2024-01-01"
        severity = "high"
        tags = "malware, keylogger, surveillance"
    
    strings:
        $keyboard_hooks = /(SetWindowsHookEx|keyboard|keystroke)/
        $input_capture = /(GetAsyncKeyState|GetKeyState|ReadConsoleInput)/
        $log_file = /(keylog|keystroke|keypress)/
        $suspicious_apis = /(WH_KEYBOARD|WH_KEYBOARD_LL)/
        
    condition:
        2 of them
}

// Правило для детекции RAT (Remote Access Trojans)
rule RAT_Detection {
    meta:
        description = "Обнаружение RAT"
        author = "CloudBridge Security Team"
        date = "2024-01-01"
        severity = "critical"
        tags = "malware, rat, remote_access"
    
    strings:
        $remote_control = /(remote|control|access|vnc|rdp)/
        $suspicious_commands = /(screenshot|webcam|microphone|file_manager)/
        $c2_communication = /(http|https|ftp|smtp)/
        $encryption = /(aes|des|rc4|blowfish)/
        
    condition:
        3 of them
}

// Правило для детекции файлов с подозрительными именами
rule Suspicious_Filenames {
    meta:
        description = "Обнаружение файлов с подозрительными именами"
        author = "CloudBridge Security Team"
        date = "2024-01-01"
        severity = "medium"
        tags = "malware, filename, evasion"
    
    strings:
        $suspicious_names = /(backdoor|trojan|virus|malware|hack|exploit)/
        $hidden_files = /^\..*$/
        $system_mimic = /(svchost|lsass|winlogon|csrss)/
        $random_names = /[a-f0-9]{32}/
        
    condition:
        any of them
}

// Правило для детекции обфусцированного кода
rule Obfuscated_Code {
    meta:
        description = "Обнаружение обфусцированного кода"
        author = "CloudBridge Security Team"
        date = "2024-01-01"
        severity = "high"
        tags = "malware, obfuscation, evasion"
    
    strings:
        $encoded_strings = /[A-Za-z0-9+/]{20,}={0,2}/
        $hex_strings = /\\x[0-9a-fA-F]{2}/
        $long_strings = /[A-Za-z0-9]{50,}/
        $suspicious_chars = /[^A-Za-z0-9\s]{20,}/
        
    condition:
        2 of them
}

// Правило для детекции persistence механизмов
rule Persistence_Detection {
    meta:
        description = "Обнаружение механизмов persistence"
        author = "CloudBridge Security Team"
        date = "2024-01-01"
        severity = "high"
        tags = "malware, persistence, startup"
    
    strings:
        $startup_locations = /(startup|autostart|run|init)/
        $registry_keys = /(HKEY_CURRENT_USER|HKEY_LOCAL_MACHINE)/
        $scheduled_tasks = /(schtasks|at|cron)/
        $service_creation = /(CreateService|sc.exe)/
        
    condition:
        2 of them
}

// Правило для детекции network scanning
rule Network_Scanner {
    meta:
        description = "Обнаружение сетевых сканеров"
        author = "CloudBridge Security Team"
        date = "2024-01-01"
        severity = "medium"
        tags = "malware, network, reconnaissance"
    
    strings:
        $scanning_tools = /(nmap|masscan|zmap|angryip)/
        $port_scanning = /(port|scan|probe|ping)/
        $network_enumeration = /(arp|netstat|ifconfig|ipconfig)/
        $suspicious_ports = /(21|22|23|25|53|80|443|445|3389)/
        
    condition:
        2 of them
}

// Правило для детекции data exfiltration
rule Data_Exfiltration {
    meta:
        description = "Обнаружение утечки данных"
        author = "CloudBridge Security Team"
        date = "2024-01-01"
        severity = "critical"
        tags = "malware, exfiltration, data_theft"
    
    strings:
        $data_collection = /(documents|desktop|downloads|pictures)/
        $file_operations = /(copy|move|upload|download)/
        $compression = /(zip|rar|7z|tar|gz)/
        $encryption = /(encrypt|password|key)/
        
    condition:
        3 of them
}

// Правило для детекции anti-analysis techniques
rule Anti_Analysis {
    meta:
        description = "Обнаружение техник противодействия анализу"
        author = "CloudBridge Security Team"
        date = "2024-01-01"
        severity = "high"
        tags = "malware, anti_analysis, evasion"
    
    strings:
        $vm_detection = /(vmware|virtualbox|vbox|qemu)/
        $debugger_detection = /(IsDebuggerPresent|CheckRemoteDebuggerPresent)/
        $sandbox_detection = /(sandbox|analysis|debug)/
        $timing_checks = /(GetTickCount|QueryPerformanceCounter)/
        
    condition:
        2 of them
} 