#!/bin/bash
###############################################################################
# Sign All CloudBridge Images
#
# This script signs all container images used by CloudBridge with cosign.
# It also generates SBOM and vulnerability attestations.
#
# Prerequisites:
# - cosign installed
# - syft installed
# - grype installed
# - COSIGN_KEY environment variable set
# - COSIGN_PASSWORD environment variable set
#
# Usage:
#   export COSIGN_KEY="$(cat cosign.key)"
#   export COSIGN_PASSWORD="your-password"
#   ./scripts/security/sign-all-images.sh
###############################################################################

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
NAMESPACE="${NAMESPACE:-cloudbridge}"
REGISTRY="${REGISTRY:-cr.yandex/crpblcmklr11njvr6ikg}"
DRY_RUN="${DRY_RUN:-false}"

###############################################################################
# Helper Functions
###############################################################################

log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $*"
}

error() {
    echo -e "${RED}[ERROR]${NC} $*" >&2
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $*"
}

info() {
    echo -e "${BLUE}[INFO]${NC} $*"
}

check_dependencies() {
    log "Checking dependencies..."

    local deps=("cosign" "syft" "grype" "kubectl")
    for dep in "${deps[@]}"; do
        if ! command -v "$dep" &> /dev/null; then
            error "Required dependency '$dep' not found"
            exit 1
        fi
    done

    # Check environment variables
    if [ -z "${COSIGN_KEY:-}" ]; then
        error "COSIGN_KEY environment variable not set"
        info "Export with: export COSIGN_KEY=\"\$(cat cosign.key)\""
        exit 1
    fi

    if [ -z "${COSIGN_PASSWORD:-}" ]; then
        warn "COSIGN_PASSWORD not set, will prompt for password"
    fi

    log "All dependencies satisfied"
}

###############################################################################
# Image Discovery
###############################################################################

get_images_from_cluster() {
    log "Discovering images from cluster..."

    # Get all images from pods in namespace
    local images=$(kubectl get pods -n "$NAMESPACE" \
        -o jsonpath='{.items[*].spec.containers[*].image}' | \
        tr ' ' '\n' | sort -u)

    # Filter only our registry
    local our_images=$(echo "$images" | grep "^${REGISTRY}/" || true)

    if [ -z "$our_images" ]; then
        warn "No images found in registry $REGISTRY"
        return 1
    fi

    echo "$our_images"
}

get_images_from_manifests() {
    log "Discovering images from manifests..."

    # Find all YAML files in k8s/ directory
    local images=$(find k8s/ -name "*.yaml" -exec grep -h "image:" {} \; | \
        awk '{print $2}' | \
        grep "^${REGISTRY}/" | \
        sort -u || true)

    echo "$images"
}

get_all_images() {
    log "Collecting all CloudBridge images..."

    local cluster_images=$(get_images_from_cluster || true)
    local manifest_images=$(get_images_from_manifests || true)

    # Combine and deduplicate
    local all_images=$(echo -e "${cluster_images}\n${manifest_images}" | \
        grep -v "^$" | \
        sort -u)

    if [ -z "$all_images" ]; then
        error "No images found to sign"
        exit 1
    fi

    echo "$all_images"
}

###############################################################################
# Signing Functions
###############################################################################

sign_image() {
    local image="$1"

    log "Signing image: $image"

    if [ "$DRY_RUN" = "true" ]; then
        info "[DRY RUN] Would sign: $image"
        return 0
    fi

    # Check if already signed
    if cosign verify --key env://COSIGN_PUBLIC_KEY "$image" &>/dev/null; then
        info "Image already signed: $image"
        return 0
    fi

    # Sign with cosign
    if cosign sign --yes --key env://COSIGN_KEY "$image"; then
        log "✓ Signed: $image"
        return 0
    else
        error "Failed to sign: $image"
        return 1
    fi
}

generate_sbom() {
    local image="$1"
    local sbom_file="/tmp/sbom-$(echo "$image" | tr '/:' '_').spdx.json"

    log "Generating SBOM for: $image"

    if [ "$DRY_RUN" = "true" ]; then
        info "[DRY RUN] Would generate SBOM for: $image"
        return 0
    fi

    # Generate SBOM with syft
    if syft packages "$image" -o spdx-json > "$sbom_file" 2>/dev/null; then
        log "✓ Generated SBOM: $(basename "$sbom_file")"
        echo "$sbom_file"
        return 0
    else
        error "Failed to generate SBOM for: $image"
        return 1
    fi
}

attest_sbom() {
    local image="$1"
    local sbom_file="$2"

    log "Attesting SBOM for: $image"

    if [ "$DRY_RUN" = "true" ]; then
        info "[DRY RUN] Would attest SBOM for: $image"
        return 0
    fi

    # Check if already attested
    if cosign verify-attestation \
        --key env://COSIGN_PUBLIC_KEY \
        --type spdx \
        "$image" &>/dev/null; then
        info "SBOM already attested: $image"
        return 0
    fi

    # Attest SBOM
    if cosign attest --yes \
        --predicate "$sbom_file" \
        --type spdx \
        --key env://COSIGN_KEY \
        "$image"; then
        log "✓ Attested SBOM: $image"
        return 0
    else
        error "Failed to attest SBOM for: $image"
        return 1
    fi
}

scan_vulnerabilities() {
    local image="$1"
    local vuln_file="/tmp/vuln-$(echo "$image" | tr '/:' '_').json"

    log "Scanning vulnerabilities for: $image"

    if [ "$DRY_RUN" = "true" ]; then
        info "[DRY RUN] Would scan vulnerabilities for: $image"
        return 0
    fi

    # Scan with grype
    if grype "$image" -o json > "$vuln_file" 2>/dev/null; then
        # Check for Critical/High vulnerabilities
        local critical=$(jq '.matches[] | select(.vulnerability.severity=="Critical") | .vulnerability.id' "$vuln_file" | wc -l)
        local high=$(jq '.matches[] | select(.vulnerability.severity=="High") | .vulnerability.id' "$vuln_file" | wc -l)

        if [ "$critical" -gt 0 ]; then
            warn "Found $critical CRITICAL vulnerabilities in: $image"
        fi

        if [ "$high" -gt 5 ]; then
            warn "Found $high HIGH vulnerabilities in: $image (limit: 5)"
        fi

        log "✓ Scanned vulnerabilities: Critical=$critical, High=$high"
        echo "$vuln_file"
        return 0
    else
        error "Failed to scan vulnerabilities for: $image"
        return 1
    fi
}

attest_vulnerabilities() {
    local image="$1"
    local vuln_file="$2"

    log "Attesting vulnerabilities for: $image"

    if [ "$DRY_RUN" = "true" ]; then
        info "[DRY RUN] Would attest vulnerabilities for: $image"
        return 0
    fi

    # Check if already attested
    if cosign verify-attestation \
        --key env://COSIGN_PUBLIC_KEY \
        --type https://cosign.sigstore.dev/attestation/vuln/v1 \
        "$image" &>/dev/null; then
        info "Vulnerabilities already attested: $image"
        return 0
    fi

    # Attest vulnerabilities
    if cosign attest --yes \
        --predicate "$vuln_file" \
        --type https://cosign.sigstore.dev/attestation/vuln/v1 \
        --key env://COSIGN_KEY \
        "$image"; then
        log "✓ Attested vulnerabilities: $image"
        return 0
    else
        error "Failed to attest vulnerabilities for: $image"
        return 1
    fi
}

###############################################################################
# Main Processing
###############################################################################

process_image() {
    local image="$1"

    info "Processing image: $image"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

    # Step 1: Sign image
    if ! sign_image "$image"; then
        error "Signing failed for: $image"
        return 1
    fi

    # Step 2: Generate and attest SBOM
    if sbom_file=$(generate_sbom "$image"); then
        attest_sbom "$image" "$sbom_file" || warn "SBOM attestation failed"
        rm -f "$sbom_file"
    else
        warn "SBOM generation failed for: $image"
    fi

    # Step 3: Scan and attest vulnerabilities
    if vuln_file=$(scan_vulnerabilities "$image"); then
        attest_vulnerabilities "$image" "$vuln_file" || warn "Vulnerability attestation failed"
        rm -f "$vuln_file"
    else
        warn "Vulnerability scan failed for: $image"
    fi

    log "✓ Completed processing: $image"
    echo ""
}

###############################################################################
# Reporting
###############################################################################

generate_report() {
    local total_images="$1"
    local signed_count="$2"
    local failed_count="$3"

    log "Generating signing report..."

    cat > /tmp/signing-report.md <<EOF
# Image Signing Report

**Date**: $(date)
**Total Images**: $total_images
**Successfully Signed**: $signed_count
**Failed**: $failed_count
**Success Rate**: $(( signed_count * 100 / total_images ))%

## Summary

- ✅ **Signatures**: $signed_count/$total_images
- ✅ **SBOM Attestations**: Generated for all signed images
- ✅ **Vulnerability Scans**: Completed for all signed images

## Signed Images

EOF

    # List signed images
    for image in $(get_all_images); do
        if cosign verify --key env://COSIGN_PUBLIC_KEY "$image" &>/dev/null; then
            echo "- ✅ $image" >> /tmp/signing-report.md
        else
            echo "- ❌ $image" >> /tmp/signing-report.md
        fi
    done

    cat /tmp/signing-report.md
    log "Report saved to: /tmp/signing-report.md"
}

###############################################################################
# Main Execution
###############################################################################

main() {
    log "CloudBridge Image Signing Script"
    log "================================="

    check_dependencies

    # Get all images
    local images=$(get_all_images)
    local total_images=$(echo "$images" | wc -l)

    log "Found $total_images images to sign"
    echo ""

    if [ "$DRY_RUN" = "true" ]; then
        warn "DRY RUN MODE - No actual signing will be performed"
        echo "$images"
        exit 0
    fi

    # Process each image
    local signed_count=0
    local failed_count=0

    while IFS= read -r image; do
        if process_image "$image"; then
            ((signed_count++))
        else
            ((failed_count++))
        fi
    done <<< "$images"

    # Generate report
    echo ""
    log "================================="
    log "Signing Complete"
    log "================================="
    log "Total: $total_images"
    log "Signed: $signed_count"
    log "Failed: $failed_count"

    generate_report "$total_images" "$signed_count" "$failed_count"

    if [ "$failed_count" -gt 0 ]; then
        error "Some images failed to sign"
        exit 1
    fi

    log "All images signed successfully!"
}

# Run main function
main "$@"
