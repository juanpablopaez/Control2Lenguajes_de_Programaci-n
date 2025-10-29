# Ejecución especulativa vs. secuencial (Go)

Este proyecto compara el rendimiento de una ejecución especulativa contra una ejecución secuencial usando goroutines y canales. La decisión de rama se toma con la traza del producto de matrices aleatorias; luego se ejecuta la rama A (Proof-of-Work) o B (primos). En modo especulativo se lanzan ambas en paralelo y se cancela cooperativamente la perdedora.

## Archivos principales
- `main.go`: programa con modos `spec`, `seq`, `bench`; CSV de métricas; cancelación cooperativa integrada.

## Compilar

1) Abrir PowerShell en la carpeta del proyecto (por ejemplo `c:\Users\juamp\OneDrive\Desktop\tarea2`).
2) Compilar:

```powershell
go build -o spec_exec.exe
```

## Ejecutar ejemplos

Ejemplo rápido (especulativo):

```powershell
./spec_exec.exe -modo spec -n 120 -umbral 2000 -powData hola -powDiff 3 -primosMax 20000 -archivo metricas.csv
```

Benchmark (30 corridas por estrategia) con parámetros pesados del enunciado:

```powershell
./spec_exec.exe -modo bench -runs 30 -n 300 -umbral 1000 -powData bloque-demo -powDiff 5 -primosMax 500000 -archivo metricas.csv
```

El modo bench escribe ambas series (`spec` y `seq`) en `metricas.csv` y muestra los promedios y el speedup en consola.

## Métricas y Speedup

El CSV contiene por corrida: `timestamp, modo, n, umbral, powDiff, primosMax, chosen, trace, total_ms` y, para especulativo, métricas por rama (`a_ok, a_ms, b_ok, b_ms`).

Speedup:

$$\text{Speedup} = \frac{T_{\text{secuencial}}}{T_{\text{especulativo}}}$$

## Reporte simple

| Estrategia   | Tiempo promedio |
|--------------|------------------|
| Especulativa |   2.35257533s    |
| Secuencial   |   2.395193843s   |

Speedup: 1.018

## Reporte completo

Vea `performance_report.md` para un reporte más extenso con gráficas y conclusiones.

## Script de análisis y gráficas

Hay un script en `tools/plot_metrics.py` que calcula promedios y genera un gráfico de speedup desde `metricas.csv`.

Requisitos (Python 3):

```powershell
python -m pip install -r requirements.txt
```

Ejemplo de uso:

```powershell
python .\tools\plot_metrics.py --csv metricas.csv --n 300 --umbral 1000 --powDiff 5 --primosMax 500000 --out graficos/speedup.png
```

El script imprimirá una tabla con promedios para `spec` y `seq`, el speedup y guardará `speedup.png`.
