Write-Host "MailRaven Environment Check" -ForegroundColor Green

# Check Config
if (Test-Path "bin\mailraven.exe") {
    Write-Host "Validating config..."
    .\bin\mailraven.exe check-config
} else {
    Write-Host "Binary not found, skipping config check." -ForegroundColor Yellow
}

Write-Host "Check complete."
