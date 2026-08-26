# build.ps1 - Build Bee Compiler
Write-Host "Building Bee Compiler..."
go build -o bee.exe .\cmd\bee\main.go
if ($LASTEXITCODE -eq 0) {
    Write-Host "Build successful: .\bee.exe" -ForegroundColor Green
} else {
    Write-Host "Build failed." -ForegroundColor Red
    exit 1
}
