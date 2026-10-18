# Reproduce los candidatos de sonido uno tras otro e imprime el nombre de cada uno.
# Uso (desde la raiz del repo):
#   powershell -ExecutionPolicy Bypass -File sound\audition\escuchar.ps1            # todo
#   powershell -ExecutionPolicy Bypass -File sound\audition\escuchar.ps1 -Cue music_menu
# Los .wav son copias de escucha: efectos completos y 30 s de las musicas (_30s).
param([string]$Cue = "")
$root = $PSScriptRoot
$dirs = Get-ChildItem -Directory $root | Where-Object { $Cue -eq "" -or $_.Name -eq $Cue }
foreach ($d in $dirs) {
    Write-Host ""
    Write-Host "=== $($d.Name) ===" -ForegroundColor Cyan
    foreach ($f in Get-ChildItem -Path $d.FullName -Filter *.wav | Sort-Object Name) {
        Write-Host ("  " + $f.Name)
        $p = New-Object System.Media.SoundPlayer $f.FullName
        $p.PlaySync()
        Start-Sleep -Milliseconds 400
    }
}
Write-Host ""
Write-Host "Listo." -ForegroundColor Green
