# GOLAP – Seminární Práce (LaTeX)

Kompletní dokumentace projektu GOLAP v LaTeX formátu podle UJEP šablony.

## 📄 Obsah dokumentu

| Sekce | Staves | Obsah |
|-------|--------|-------|
| Titulní strana | 1 | Údaje autora, datum, název práce |
| Obsah | 2–3 | Tabulka obsahu (auto-generuje se) |
| **1. Úvod** | 3 | Cíle práce, motivation |
| **2. Volba nástroje** | 3–4 | Srovnění DuckDB vs PostgreSQL |
| **3. Instalace a prostředí** | 4 | Setup projektu, dependencies, struktura |
| **4. Datová sada** | 5 | Dataset (550K záznamů), charakteristika |
| **5. Datová struktura** | 5–6 | Star schema PostgreSQL, flat DuckDB |
| **6. OLAP řezy** | 7–12 | 6 pokročilých queries (Q1–Q6) |
|   - Q1: ROLLUP | 7 | Hierarchie dekád → let → modely |
|   - Q2: CUBE | 8 | Model × Karoserie × Převodovka |
|   - Q3: GROUPING SETS | 8–9 | Barvy u top modelů |
|   - Q4: ROLLUP (heatmapa) | 9 | Značka × Karoserie |
|   - Q5: CUBE (segmentace) | 9–10 | Model × Převodovka × Cena |
|   - Q6: GROUPING SETS (multi) | 10–11 | 4 analytické pohledy |
| **7. Benchmark** | 11–12 | Výkonnostní srovnání DuckDB vs PG |
| **8. Data Mining** | 12–14 | K-Means, regrese, anomálie |
| **9. Závěr** | 14 | Dosažené cíle, technical insights |
| **10. Zdroje** | 15 | Citace (ISO 690) |

**Přibližně 15–18 stran A4** s grafy a tabulkami.

---

## 🔧 Jak pracovat s dokumentem

### 1. Editace jména a údajů

Vše relevantní je v **preambulovém kódu** (řádky 1–70):

```latex
\newcommand{\modsinfo}{%
    \begin{flushright}
        Autor: Matěj Bureš  % ← ZMĚŇ ZDE
        ...
    \end{flushright}
}
```

### 2. Přidání grafů

Každá sekce OLAP queries má místo pro grafy. Pokud budeš mít obrázky, přidej je takto:

```latex
\subsection{Q1: Hierarchie...}
...
\begin{figure}[H]
    \centering
    \includegraphics[width=0.9\textwidth]{../output/q1_trend.png}
    \caption{Q1: Prodeje v závislosti na roku výroby}
    \label{fig:q1}
\end{figure}
```

**Kde najít obrázky:**
- Vygeneruješ je z `frontend/frontend.ipynb` (Jupyter notebook)
- Ulož je do `output/` nebo `documentation/figures/`
- Pak je referencuj relativní cestou (např. `../output/q1_trend.png`)

### 3. Přidání benchmark grafu

Benchmarky jsou v tabulce, ale můžeš přidat vizuální graf:

```latex
\subsection{Vizualizace benchmarků}
\begin{figure}[H]
    \centering
    \includegraphics[width=\textwidth]{../output/benchmark_comparison.png}
    \caption{Srovnění časů zpracování}
\end{figure}
```

### 4. Editace SQL kódů

Všechny SQL příkazy jsou v `\begin{lstlisting}...\end{lstlisting}` blocích.
Pokud najdeš chybu nebo chceš updatovat query, jednoduše edituj text mezi těmito tagy.

---

## 📦 Kompilace do PDF

### Varianty:

#### **Možnost 1: Online (Overleaf.com)**
1. Vytvoř nový projekt na https://www.overleaf.com
2. Zkopíruj obsah `GOLAP_seminární_práce.tex` do Overleaf
3. Klikni "Recompile" → PDF se vygeneruje

#### **Možnost 2: Lokálně (LaTeX na počítači)**

Potřebuješ:
- **MiKTeX** (Windows): https://miktex.org/download
- **TeX Live** (Linux/Mac): https://www.tug.org/texlive/

Pak v terminálu (v adresáři `documentation/`):

```bash
pdflatex -interaction=nonstopmode GOLAP_seminární_práce.tex
bibtex GOLAP_seminární_práce  # (pokud máš .bib soubor)
pdflatex -interaction=nonstopmode GOLAP_seminární_práce.tex  # 2x
\end{bash}
```

Nebo jednoduš:
```bash
latexmk -pdf GOLAP_seminární_práce.tex
```

---

## 🎨 Možnosti úprav

### Změna jazyka
Pokud chceš angličtinu místo češtiny:
```latex
\usepackage[english]{babel}  % místo [czech]
```

### Změna šíře stránky
```latex
\geometry{
    left=3cm,      % zvětシ si okraje
    right=3cm,
    ...
}
```

### Změna fontu
Přidej po `\usepackage[czech]{babel}`:
```latex
\usepackage{lmodern}          % modernější font
% nebo
\usepackage{times}            % klasické Times New Roman
```

---

## 📊 Kde se mají objevit grafy?

Grafy z notebooku mohou jít do těchto sekcí:

| Sekce | Typ grafu |
|-------|-----------|
| **Q1** | Lineplot – prodeje v čase (rok) |
| **Q2** | Barplot – průměrné ceny karoserií × převodovka |
| **Q3** | FacetGrid – barvy u top modelů |
| **Q4** | Heatmapa – Značka × Karoserie (objem vs cena) |
| **Q5** | Matrix pie charts – Model × Převodovka × Segment |
| **Q6** | Stem plot (top modely) + line plot (trendy) |
| **Benchmark** | Bar/line plot – časy DuckDB vs PostgreSQL |
| **Clustering** | Scatter plot – cena vs. nájezd (barevné objem) |
| **Regrese** | Scatter + trend line – cena vs. rok |
| **Anomálie** | Scatter – MMR vs. selling_price |

---

## ✅ Checklist před odevzdáním

- [ ] Jméno autora je správné
- [ ] Akademický rok je správný (2025/2026)
- [ ] Grafy z notebooku jsou exportované a zahrnuté
- [ ] SQL kódy jsou správné (zkontroluj s Go souborů)
- [ ] Benchmark tabulka odpovídá reálným měřením
- [ ] Všechny citace jsou ve formátu ISO 690
- [ ] PDF je napsáno bez chyb (zkompiluj si ho!)

---

## 🚀 Další kroky

1. **Vygeneruj grafy z notebooku:**
   - Spusť `frontend/frontend.ipynb` v Jupyter
   - Exportuj všechny visibility grafy jako PNG (png parametr v seaborn)
   - Ulož je do `documentation/figures/` nebo `output/`

2. **Updatuj benchmark data:**
   - Spusť Go main.go s benchmark flagy
   - Zkopíruj časy do tabulky v sekci 7 (Benchmark)

3. **Zkompiluj PDF:**
   - Vyzkoušej MiKTeX, TeX Live nebo Overleaf
   - Ověř, že se vše správně zformátovalo

4. **Odevzd:**
   - PDF a .tex soubor vedoucímu
   - Případně i surové .csv s výsledky queries

---

## 📚 Užitečné linky

- **UJEP šablony**: https://www.ujep.cz
- **Overleaf (online LaTeX editor)**: https://www.overleaf.com
- **DuckDB dokumentace**: https://duckdb.org/docs/
- **PostgreSQL OLAP**: https://www.postgresql.org/docs/current/queries-group.html
- **ISO 690 citace**: https://www.citace.com/

---

Hodně štěstí s prací! 🎓
