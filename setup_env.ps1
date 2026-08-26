# setup_env.ps1
$beePath = Join-Path (Get-Location).Path "bin"
$env:Path += ";$beePath"
Write-Host "Bee compiler added to current session PATH: $beePath"
