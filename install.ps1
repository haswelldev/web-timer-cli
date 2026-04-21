$arch = if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") { "arm64" } else { "amd64" }
$url = "https://github.com/haswelldev/web-timer-cli/releases/latest/download/web-timer-cli-windows-$arch.exe"
$dest = "$env:USERPROFILE\AppData\Local\Microsoft\WindowsApps\web-timer-cli.exe"
Invoke-WebRequest $url -OutFile $dest
Write-Host "Installed to $dest"
Write-Host "To uninstall: Remove-Item $dest"
