# setup_env.ps1
$beePath = Join-Path (Get-Location).Path "bin"
$existingPath = [Environment]::GetEnvironmentVariable("Path", "User")
if (-not $existingPath.Contains($beePath)) {
    [Environment]::SetEnvironmentVariable("Path", $existingPath + ";" + $beePath, "User")
    Write-Host "Bee compiler permanently added to User PATH: $beePath"
} else {
    Write-Host "Bee compiler already in User PATH: $beePath"
}
