param(
  [int]$runs = 30,
  [int]$n = 300,
  [int]$umbral = 1000,
  [string]$powData = "bloque-demo",
  [int]$powDiff = 5,
  [int]$primosMax = 500000,
  [string]$archivo = "metricas.csv"
)

Write-Host "== Ejecutando benchmark =="
Write-Host "runs=$runs n=$n umbral=$umbral powData=$powData powDiff=$powDiff primosMax=$primosMax archivo=$archivo"

./spec_exec.exe -modo bench -runs $runs -n $n -umbral $umbral -powData $powData -powDiff $powDiff -primosMax $primosMax -archivo $archivo

Write-Host "Listo. Resultados en $archivo"