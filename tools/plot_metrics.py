import argparse
import sys
from pathlib import Path

import pandas as pd
import matplotlib.pyplot as plt


def load_and_filter(csv_path: Path, n: int, umbral: int, powDiff: int, primosMax: int) -> pd.DataFrame:
    df = pd.read_csv(csv_path)
    # Normalizar nombres esperados
    required = {
        'timestamp', 'modo', 'n', 'umbral', 'powDiff', 'primosMax',
        'chosen', 'trace', 'total_ms'
    }
    missing = [c for c in required if c not in df.columns]
    if missing:
        raise SystemExit(f"Faltan columnas en CSV: {missing}")
    # Filtrar por parámetros
    mask = (
        (df['n'] == n) &
        (df['umbral'] == umbral) &
        (df['powDiff'] == powDiff) &
        (df['primosMax'] == primosMax)
    )
    return df.loc[mask].copy()


def compute_means(df: pd.DataFrame) -> pd.DataFrame:
    # Asegurar tipo numérico
    df['total_ms'] = pd.to_numeric(df['total_ms'], errors='coerce')
    grouped = df.groupby('modo', as_index=False)['total_ms'].mean().rename(columns={'total_ms': 'avg_ms'})
    return grouped


def plot_speedup(means: pd.DataFrame, out: Path):
    # Esperamos filas para 'spec' y 'seq'
    modes = ['spec', 'seq']
    m = {row['modo']: row['avg_ms'] for _, row in means.iterrows()}
    if not all(k in m for k in modes):
        print("Advertencia: no hay datos completos de ambos modos para graficar.")
        return
    speedup = m['seq'] / m['spec'] if m['spec'] > 0 else float('inf')

    fig, ax = plt.subplots(figsize=(5, 4))
    ax.bar(['Especulativo', 'Secuencial'], [m['spec'], m['seq']], color=['#2ca02c', '#1f77b4'])
    ax.set_ylabel('Tiempo promedio (ms)')
    ax.set_title(f'Speedup = Tseq/Tspec = {speedup:.3f}')
    for i, v in enumerate([m['spec'], m['seq']]):
        ax.text(i, v, f"{v:.1f}", ha='center', va='bottom')
    fig.tight_layout()
    fig.savefig(out, dpi=150)
    print(f"Grafico guardado en: {out}")


def main():
    parser = argparse.ArgumentParser(description='Analiza metricas.csv y grafica speedup')
    parser.add_argument('--csv', required=True, help='Ruta a metricas.csv')
    parser.add_argument('--n', type=int, required=True)
    parser.add_argument('--umbral', type=int, required=True)
    parser.add_argument('--powDiff', type=int, required=True)
    parser.add_argument('--primosMax', type=int, required=True)
    parser.add_argument('--out', default='speedup.png', help='Archivo PNG de salida')
    args = parser.parse_args()

    csv_path = Path(args.csv)
    if not csv_path.exists():
        print(f"No existe CSV: {csv_path}")
        sys.exit(1)

    df = load_and_filter(csv_path, args.n, args.umbral, args.powDiff, args.primosMax)
    if df.empty:
        print("No hay filas que coincidan con los parámetros dados.")
        sys.exit(2)

    means = compute_means(df)
    print('\nPromedios por modo (ms):')
    print(means.to_string(index=False))

    # Calcular speedup si están ambos modos
    try:
        spec_avg = float(means.loc[means['modo'] == 'spec', 'avg_ms'].iloc[0])
        seq_avg = float(means.loc[means['modo'] == 'seq', 'avg_ms'].iloc[0])
        speedup = seq_avg / spec_avg if spec_avg > 0 else float('inf')
        print(f"\nSpeedup = Tseq/Tspec = {speedup:.3f}")
    except IndexError:
        print("\nAdvertencia: faltan datos de alguno de los modos para calcular speedup.")

    plot_speedup(means, Path(args.out))


if __name__ == '__main__':
    main()
