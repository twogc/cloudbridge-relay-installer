#!/bin/bash

# CloudBridge Relay Log Management Script
# This script provides utilities for managing application logs

set -e

# Configuration
LOG_DIR="${LOG_DIR:-./logs}"
LOG_FILE="${LOG_FILE:-relay.log}"
MAX_LOG_SIZE="${MAX_LOG_SIZE:-100M}"
MAX_LOG_AGE="${MAX_LOG_AGE:-30d}"
COMPRESS_LOGS="${COMPRESS_LOGS:-true}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Log levels
ERROR=0
WARN=1
INFO=2
DEBUG=3

# Current log level
LOG_LEVEL=${LOG_LEVEL:-$INFO}

# Logging function
log() {
    local level=$1
    shift
    local message="$*"
    
    if [ $level -le $LOG_LEVEL ]; then
        case $level in
            $ERROR) echo -e "${RED}[ERROR]${NC} $message" >&2 ;;
            $WARN)  echo -e "${YELLOW}[WARN]${NC} $message" ;;
            $INFO)  echo -e "${GREEN}[INFO]${NC} $message" ;;
            $DEBUG) echo -e "${BLUE}[DEBUG]${NC} $message" ;;
        esac
    fi
}

# Show usage
show_usage() {
    cat << EOF
CloudBridge Relay Log Management Script

Usage: $0 [COMMAND] [OPTIONS]

Commands:
    rotate      Rotate log files
    clean       Clean old log files
    analyze     Analyze log files for patterns
    tail        Tail log files in real-time
    search      Search for specific patterns in logs
    stats       Show log statistics
    compress    Compress old log files
    setup       Setup log rotation configuration

Options:
    -d, --dir DIR       Log directory (default: $LOG_DIR)
    -f, --file FILE     Log file name (default: $LOG_FILE)
    -s, --size SIZE     Max log size (default: $MAX_LOG_SIZE)
    -a, --age AGE       Max log age (default: $MAX_LOG_AGE)
    -c, --compress      Compress logs (default: $COMPRESS_LOGS)
    -v, --verbose       Verbose output
    -h, --help          Show this help

Examples:
    $0 rotate                    # Rotate logs
    $0 clean --age 7d           # Clean logs older than 7 days
    $0 analyze --file error.log # Analyze error logs
    $0 tail --follow            # Follow log in real-time
    $0 search "ERROR"           # Search for ERROR patterns
    $0 stats                    # Show log statistics

EOF
}

# Check if required tools are available
check_dependencies() {
    local missing_tools=()
    
    for tool in jq awk sed grep find; do
        if ! command -v $tool &> /dev/null; then
            missing_tools+=($tool)
        fi
    done
    
    if [ ${#missing_tools[@]} -ne 0 ]; then
        log $ERROR "Missing required tools: ${missing_tools[*]}"
        exit 1
    fi
}

# Create log directory if it doesn't exist
ensure_log_dir() {
    if [ ! -d "$LOG_DIR" ]; then
        log $INFO "Creating log directory: $LOG_DIR"
        mkdir -p "$LOG_DIR"
    fi
}

# Rotate log files
rotate_logs() {
    log $INFO "Rotating log files in $LOG_DIR"
    
    local log_path="$LOG_DIR/$LOG_FILE"
    
    if [ ! -f "$log_path" ]; then
        log $WARN "Log file not found: $log_path"
        return 0
    fi
    
    # Get current log size
    local current_size=$(stat -c%s "$log_path" 2>/dev/null || stat -f%z "$log_path" 2>/dev/null || echo 0)
    local max_size_bytes=$(numfmt --from=iec "$MAX_LOG_SIZE")
    
    if [ $current_size -gt $max_size_bytes ]; then
        local timestamp=$(date +%Y%m%d_%H%M%S)
        local rotated_file="${log_path}.${timestamp}"
        
        log $INFO "Rotating log file: $log_path -> $rotated_file"
        mv "$log_path" "$rotated_file"
        
        # Create new log file
        touch "$log_path"
        chmod 644 "$log_path"
        
        # Compress if enabled
        if [ "$COMPRESS_LOGS" = "true" ]; then
            log $INFO "Compressing rotated log: $rotated_file"
            gzip "$rotated_file" &
        fi
        
        log $INFO "Log rotation completed"
    else
        log $DEBUG "Log file size ($current_size bytes) is below threshold ($max_size_bytes bytes)"
    fi
}

# Clean old log files
clean_logs() {
    log $INFO "Cleaning old log files in $LOG_DIR"
    
    local deleted_count=0
    local freed_space=0
    
    # Find and delete old log files
    while IFS= read -r -d '' file; do
        local file_size=$(stat -c%s "$file" 2>/dev/null || stat -f%z "$file" 2>/dev/null || echo 0)
        log $INFO "Deleting old log file: $file (${file_size} bytes)"
        rm -f "$file"
        deleted_count=$((deleted_count + 1))
        freed_space=$((freed_space + file_size))
    done < <(find "$LOG_DIR" -name "*.log.*" -type f -mtime +${MAX_LOG_AGE%[dhms]} -print0 2>/dev/null)
    
    # Find and delete old compressed files
    while IFS= read -r -d '' file; do
        local file_size=$(stat -c%s "$file" 2>/dev/null || stat -f%z "$file" 2>/dev/null || echo 0)
        log $INFO "Deleting old compressed log: $file (${file_size} bytes)"
        rm -f "$file"
        deleted_count=$((deleted_count + 1))
        freed_space=$((freed_space + file_size))
    done < <(find "$LOG_DIR" -name "*.log.*.gz" -type f -mtime +${MAX_LOG_AGE%[dhms]} -print0 2>/dev/null)
    
    if [ $deleted_count -gt 0 ]; then
        log $INFO "Cleaned $deleted_count files, freed $(numfmt --to=iec $freed_space) bytes"
    else
        log $INFO "No old log files found to clean"
    fi
}

# Analyze log files
analyze_logs() {
    local log_file="$1"
    [ -z "$log_file" ] && log_file="$LOG_DIR/$LOG_FILE"
    
    if [ ! -f "$log_file" ]; then
        log $ERROR "Log file not found: $log_file"
        return 1
    fi
    
    log $INFO "Analyzing log file: $log_file"
    
    # Basic statistics
    local total_lines=$(wc -l < "$log_file")
    local file_size=$(stat -c%s "$log_file" 2>/dev/null || stat -f%z "$log_file" 2>/dev/null || echo 0)
    
    echo "=== Log Analysis Report ==="
    echo "File: $log_file"
    echo "Size: $(numfmt --to=iec $file_size)"
    echo "Lines: $total_lines"
    echo ""
    
    # Error analysis
    echo "=== Error Analysis ==="
    local error_count=$(grep -c "ERROR\|CRITICAL" "$log_file" || echo 0)
    local warn_count=$(grep -c "WARN\|WARNING" "$log_file" || echo 0)
    local info_count=$(grep -c "INFO" "$log_file" || echo 0)
    
    echo "Errors: $error_count"
    echo "Warnings: $warn_count"
    echo "Info messages: $info_count"
    echo ""
    
    # Top error patterns
    echo "=== Top Error Patterns ==="
    grep "ERROR\|CRITICAL" "$log_file" | sed 's/.*ERROR[[:space:]]*//' | sort | uniq -c | sort -nr | head -10
    echo ""
    
    # Time-based analysis
    echo "=== Time-based Analysis ==="
    echo "Log entries by hour:"
    grep -o '[0-9]\{4\}-[0-9]\{2\}-[0-9]\{2\}T[0-9]\{2\}:' "$log_file" | sort | uniq -c | sort -k2
    echo ""
    
    # Performance analysis
    echo "=== Performance Analysis ==="
    echo "Slow operations (>1s):"
    grep "duration.*[0-9]\{2,\}ms\|duration.*[0-9]\.[0-9]s" "$log_file" | tail -5
    echo ""
}

# Tail log files
tail_logs() {
    local log_file="$1"
    [ -z "$log_file" ] && log_file="$LOG_DIR/$LOG_FILE"
    
    if [ ! -f "$log_file" ]; then
        log $ERROR "Log file not found: $log_file"
        return 1
    fi
    
    log $INFO "Tailing log file: $log_file"
    
    # Check if follow option is provided
    if [[ "$*" == *"--follow"* ]] || [[ "$*" == *"-f"* ]]; then
        tail -f "$log_file"
    else
        tail -n 50 "$log_file"
    fi
}

# Search logs
search_logs() {
    local pattern="$1"
    local log_file="$2"
    [ -z "$log_file" ] && log_file="$LOG_DIR/$LOG_FILE"
    
    if [ -z "$pattern" ]; then
        log $ERROR "Search pattern is required"
        return 1
    fi
    
    if [ ! -f "$log_file" ]; then
        log $ERROR "Log file not found: $log_file"
        return 1
    fi
    
    log $INFO "Searching for pattern '$pattern' in $log_file"
    
    # Search with context
    if [[ "$*" == *"--context"* ]] || [[ "$*" == *"-C"* ]]; then
        grep -C 2 -i "$pattern" "$log_file"
    else
        grep -i "$pattern" "$log_file"
    fi
}

# Show log statistics
show_stats() {
    log $INFO "Generating log statistics for $LOG_DIR"
    
    echo "=== Log Statistics ==="
    echo "Directory: $LOG_DIR"
    echo ""
    
    # File count and sizes
    local total_files=$(find "$LOG_DIR" -name "*.log*" -type f | wc -l)
    local total_size=$(find "$LOG_DIR" -name "*.log*" -type f -exec stat -c%s {} \; 2>/dev/null | awk '{sum+=$1} END {print sum+0}')
    
    echo "Total log files: $total_files"
    echo "Total size: $(numfmt --to=iec $total_size)"
    echo ""
    
    # File breakdown
    echo "File breakdown:"
    find "$LOG_DIR" -name "*.log*" -type f -exec ls -lh {} \; | awk '{print $5, $9}' | sort -k2
    echo ""
    
    # Recent activity
    echo "Recent activity (last 24 hours):"
    find "$LOG_DIR" -name "*.log*" -type f -mtime -1 -exec ls -lh {} \;
    echo ""
    
    # Disk usage
    echo "Disk usage:"
    du -sh "$LOG_DIR"
}

# Compress old log files
compress_logs() {
    log $INFO "Compressing old log files in $LOG_DIR"
    
    local compressed_count=0
    
    # Find uncompressed log files older than 1 day
    while IFS= read -r -d '' file; do
        if [[ "$file" != *.gz ]]; then
            log $INFO "Compressing: $file"
            gzip "$file"
            compressed_count=$((compressed_count + 1))
        fi
    done < <(find "$LOG_DIR" -name "*.log.*" -type f -mtime +1 -print0 2>/dev/null)
    
    if [ $compressed_count -gt 0 ]; then
        log $INFO "Compressed $compressed_count files"
    else
        log $INFO "No files to compress"
    fi
}

# Setup log rotation
setup_rotation() {
    log $INFO "Setting up log rotation configuration"
    
    # Create logrotate configuration
    local logrotate_conf="/etc/logrotate.d/cloudbridge-relay"
    
    cat > "$logrotate_conf" << EOF
$LOG_DIR/*.log {
    daily
    missingok
    rotate 30
    compress
    delaycompress
    notifempty
    create 644 root root
    postrotate
        systemctl reload cloudbridge-relay || true
    endscript
}
EOF
    
    log $INFO "Created logrotate configuration: $logrotate_conf"
    log $INFO "Log rotation will run daily and keep 30 days of logs"
}

# Main function
main() {
    # Parse command line arguments
    local command=""
    local args=()
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            -d|--dir)
                LOG_DIR="$2"
                shift 2
                ;;
            -f|--file)
                LOG_FILE="$2"
                shift 2
                ;;
            -s|--size)
                MAX_LOG_SIZE="$2"
                shift 2
                ;;
            -a|--age)
                MAX_LOG_AGE="$2"
                shift 2
                ;;
            -c|--compress)
                COMPRESS_LOGS="true"
                shift
                ;;
            -v|--verbose)
                LOG_LEVEL=$DEBUG
                shift
                ;;
            -h|--help)
                show_usage
                exit 0
                ;;
            -*)
                log $ERROR "Unknown option: $1"
                show_usage
                exit 1
                ;;
            *)
                if [ -z "$command" ]; then
                    command="$1"
                else
                    args+=("$1")
                fi
                shift
                ;;
        esac
    done
    
    # Check dependencies
    check_dependencies
    
    # Ensure log directory exists
    ensure_log_dir
    
    # Execute command
    case $command in
        rotate)
            rotate_logs
            ;;
        clean)
            clean_logs
            ;;
        analyze)
            analyze_logs "${args[@]}"
            ;;
        tail)
            tail_logs "${args[@]}"
            ;;
        search)
            search_logs "${args[@]}"
            ;;
        stats)
            show_stats
            ;;
        compress)
            compress_logs
            ;;
        setup)
            setup_rotation
            ;;
        "")
            log $ERROR "Command is required"
            show_usage
            exit 1
            ;;
        *)
            log $ERROR "Unknown command: $command"
            show_usage
            exit 1
            ;;
    esac
}

# Run main function
main "$@" 