param(
    [string]$Python = "",
    [string]$Results = "validation/results-full",
    [int]$Seed = 20260719,
    [switch]$SkipFailureFollowUp
)

$ErrorActionPreference = "Stop"
$repo = Split-Path -Parent $PSScriptRoot
Set-Location $repo

$exe = Join-Path $repo "validation\coolprop-validate.exe"
go build -o $exe ./validation/cmd/coolprop-validate

if ([string]::IsNullOrWhiteSpace($Python)) {
    $candidates = @(
        "C:\Users\Roelof Jan\AppData\Local\Programs\Python\Python313\python.exe",
        "python.exe"
    )
    foreach ($candidate in $candidates) {
        try {
            $version = & $candidate --version 2>&1
            & $candidate -c "import CoolProp" 2>$null
            if ($LASTEXITCODE -eq 0) {
                $Python = $candidate
                break
            }
        } catch {
        }
    }
}
if ([string]::IsNullOrWhiteSpace($Python)) {
    throw "No Python executable with CoolProp was found. Pass -Python explicitly."
}

New-Item -ItemType Directory -Force -Path $Results | Out-Null
$baseArgs = @(
    "-all", "-python", $Python, "-seed", $Seed,
    "-td-t", 64, "-td-rho", 64, "-pt-t", 64, "-pt-p", 64,
    "-quasi", 2048, "-results", $Results
)
& $exe @baseArgs
if ($LASTEXITCODE -ne 0) {
    throw "Initial validation failed with exit code $LASTEXITCODE."
}

$runDirs = Get-ChildItem -Directory $Results | Sort-Object Name -Descending
$latest = $runDirs | Select-Object -First 1
$failed = @(Get-ChildItem -Path (Join-Path $latest.FullName "failed") -Recurse -Filter failures.json -ErrorAction SilentlyContinue)
if ($failed.Count -eq 0 -or $SkipFailureFollowUp) {
    $index = Get-Content -Raw (Join-Path $latest.FullName "index.md")
    if ($index.Contains("planned_not_executed")) {
        Write-Error "Statistical validation is incomplete: one or more required suites are planned_not_executed. Run cannot be certified."
        exit 2
    }
    Write-Output "Validation complete: $($latest.FullName)"
    exit 0
}

Write-Output "Failures detected; running integer-count exhaustive follow-up screening."
$followUpArgs = @(
    "-all", "-python", $Python, "-seed", $Seed,
    "-td-t", 256, "-td-rho", 256, "-pt-t", 256, "-pt-p", 256,
    "-quasi", 10000, "-results", $Results
)
& $exe @followUpArgs
if ($LASTEXITCODE -ne 0) {
    throw "Integer follow-up validation failed with exit code $LASTEXITCODE."
}
$followUp = Get-ChildItem -Directory $Results | Sort-Object Name -Descending | Select-Object -First 1
$followUpIndex = Get-Content -Raw (Join-Path $followUp.FullName "index.md")
if ($followUpIndex.Contains("planned_not_executed")) {
    Write-Error "Integer follow-up completed only screening cases; required statistical suites remain planned_not_executed. Run cannot be certified."
    exit 2
}
Write-Output "Integer follow-up complete: $($followUp.FullName)"
