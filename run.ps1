
$wslIp = (wsl hostname -I).Trim().Split(" ")[0]
Write-Host "Zjištěná WSL IP: $wslIp" -ForegroundColor Cyan

# 2. Nastav ji pro Go
$env:DB_HOST = $wslIp

# 3. Spusť Go
go run ./main.go