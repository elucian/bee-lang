# test.ps1 - Run Test Pipeline
Write-Host "Running Test Pipeline..."
python test/test_runner.py
if ($LASTEXITCODE -eq 0) {
    Write-Host "Tests passed." -ForegroundColor Green
} else {
    Write-Host "Tests failed." -ForegroundColor Red
    exit 1
}
