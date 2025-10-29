# Reporte de rendimiento

Este documento presenta el estudio de rendimiento comparando ejecución especulativa vs. secuencial.

## Metodología

- Parámetros usados (ejemplo):
  - `n = 300`
  - `umbral = 0` 
  - `powDiff = 5`
  - `primosMax = 500000`
- Se ejecutaron 30 corridas especulativas y 30 secuenciales con los mismos parámetros.
- Se registró `metricas.csv` y se procesó con `tools/plot_metrics.py`.

## Resultados (complete con sus datos)

| Estrategia   | Promedio (ms)  | Speedup (ms)  |
|--------------|----------------|---------------|
| Secuencial   |  2.37117341s   | 0,989         |
| Especulativa |  2.373666176s  |               | 
| Secuencial   |  2390.484267   | 1.013         |
| Especulativa |  2358.729367   |               |
| Secuencial   |  2.997767      | 0.989         |
| Especulativa |  3.031900      |               |
| Secuencial   |  40.285300     | 1.757         |
| Especulativa |  22.925333     |               |



## Gráficas

los graficos se encuentra en la carpeta graficos.

## Observaciones y conclusiones

- Describa cómo varía el speedup con distintos parámetros (`powDiff`, `primosMax`, `n`, `umbral`).
- Explique por qué la especulación puede no ayudar si el costo de decisión es alto o si la rama perdedora termina antes de cancelar.
- Mencione el overhead observable (lanzar 2 goroutines, coordinación, cancelación cooperativa).

Para describir como varia debemos entender primero el funcionamiento de los parametros a insertar, los cuales son:
powDiff: Es el parametro que se entrega para la funcion Proof-of-Work donde este valor es la dificultad que va a tardar la funcion.
primosMax: Es el parametro que se entrega para la funcion EncontrarPrimos, el cual indica la cantidad maxima de numeros primos a encontrar.
n: Es el parametro que dicta el numero para la dimension de matrices para la desicion.
umbral: Es el parametro que decide cual es la rama ganadora donde si trace >= umbral se elije la funcion Proof-of-Work sino EncontrarPrimos.

Al comprender esto tendremos que al cambiar ciertos parametros se entiende la ejecucion del programa, para esto tendremos 4 comandos de ejecucion los cuales seran:

1)./spec_exec.exe -modo=bench -powDiff=5 -primosMax=500000 -umbral=0 -n=300 
2)./spec_exec.exe -modo=bench -powDiff=5 -primosMax=5 -umbral=0 -n=300 -archivo=metricas2.cvs
3)./spec_exec.exe -modo=bench -powDiff=5000 -primosMax=5000 -umbral=99999999 -n=300 -archivo=metricas3.cvs
4)./spec_exec.exe -modo=bench -powDiff=5000 -primosMax=100000 -umbral=99999999 -n=800 -archivo=metricas4.cvs

cada uno posteriormente se utilizara el siguiente comando de ejecucion para generar el grafico:

1)python .\tools\plot_metrics.py --csv metricas.csv --n 300 --umbral 0 --powDiff 5 --primosMax 500000 --out graficos/speedup.png
2)python .\tools\plot_metrics.py --csv metricas2 --n 300 --umbral 0 --powDiff 5 --primosMax 5 --out graficos/speedup1.png
3)python .\tools\plot_metrics.py --csv metricas3 --n 300 --umbral 99999999 --powDiff 5000 --primosMax 5000 --out graficos/speedup2.png
4)python .\tools\plot_metrics.py --csv metricas4 --n 800 --umbral 99999999 --powDiff 5000 --primosMax 100000 --out graficos/speedup3.png


Con esto podemos concluir que la ejecucion especulativa no siempre acelera un programa, debido a que en las ejecuciones 1 y 2, no se ven grandes cambios en la ejecucion, pero en la 4 ejecucion se obtuvo una mejora significativa, porque la tarea de cómputo se ejecuta mientras se toma la desicion, terminando casi al mismo tiempo y en la ejecucion 3 es tan rapida que no hay diferencia.