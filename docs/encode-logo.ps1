# Encode logo for Operator Catalog CSV (Windows PowerShell)

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$LogoFile = Join-Path $ScriptDir "logo-catalog.svg"

if (-not (Test-Path $LogoFile)) {
    Write-Error "Error: $LogoFile not found"
    exit 1
}

Write-Host "Encoding $LogoFile to base64..." -ForegroundColor Green
Write-Host ""
Write-Host "Copy the following base64 string to your ClusterServiceVersion metadata:" -ForegroundColor Yellow
Write-Host "============================================================================" -ForegroundColor Cyan
Write-Host ""

$bytes = [System.IO.File]::ReadAllBytes($LogoFile)
$base64 = [System.Convert]::ToBase64String($bytes)
Write-Host $base64

Write-Host ""
Write-Host "============================================================================" -ForegroundColor Cyan
Write-Host "Add this to your CSV file under spec.icon:" -ForegroundColor Yellow
Write-Host ""
Write-Host "  icon:" -ForegroundColor Gray
Write-Host "  - base64data: <paste-base64-here>" -ForegroundColor Gray
Write-Host "    mediatype: image/svg+xml" -ForegroundColor Gray
Write-Host ""

# Optional: Speichern in Datei
$OutputFile = Join-Path $ScriptDir "logo-catalog-base64.txt"
$base64 | Out-File -FilePath $OutputFile -Encoding ASCII
Write-Host "Base64 string also saved to: $OutputFile" -ForegroundColor Green
#!/bin/bash
# Encode logo for Operator Catalog CSV

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LOGO_FILE="$SCRIPT_DIR/logo-catalog.svg"

if [ ! -f "$LOGO_FILE" ]; then
    echo "Error: $LOGO_FILE not found"
    exit 1
fi

echo "Encoding $LOGO_FILE to base64..."
echo ""
echo "Copy the following base64 string to your ClusterServiceVersion metadata:"
echo "============================================================================"
echo ""

if command -v base64 &> /dev/null; then
    # Linux/Mac
    base64 -w 0 "$LOGO_FILE"
elif command -v openssl &> /dev/null; then
    # Alternative mit openssl
    openssl base64 -A -in "$LOGO_FILE"
else
    echo "Error: Neither 'base64' nor 'openssl' command found"
    exit 1
fi

echo ""
echo ""
echo "============================================================================"
echo "Add this to your CSV file under spec.icon:"
echo ""
echo "  icon:"
echo "  - base64data: <paste-base64-here>"
echo "    mediatype: image/svg+xml"
echo ""

