#!/usr/bin/env python3
"""Derive the report and post appendix from raw production-audit artifacts."""
import collections
import csv
import json
import pathlib
import re
import statistics

ROOT = pathlib.Path(__file__).resolve().parents[2]
OUT = ROOT / 'docs/results/production-2026-10-13'
NAMES = ['vrh-ring', 'go-rendezvous-xxhash', 'nspcc-score-scan-adapter', 'xxhash-pair-reference']
SHORT = ['VRH Ring', 'go-rendezvous + XXH64', 'NSPCC scan¹', 'XXH64 pair¹']


def load_rows(name):
    return [json.loads(line) for line in (OUT / name).read_text().splitlines()]


def parse_bench(name):
    groups = collections.defaultdict(list)
    for line in (OUT / name).read_text().splitlines():
        m = re.match(r'^(Benchmark\S+)\s+(\d+)\s+([\d.]+) ns/op\s+(\d+) B/op\s+(\d+) allocs/op$', line)
        if not m:
            continue
        label, iterations, ns, allocated, allocs = m.groups()
        groups[label].append(dict(iterations=int(iterations), ns=float(ns), bytes=int(allocated), allocs=int(allocs)))
    return {label: dict(repeats=len(rows), ns=statistics.median(r['ns'] for r in rows),
                        min_ns=min(r['ns'] for r in rows), max_ns=max(r['ns'] for r in rows),
                        bytes=statistics.median(r['bytes'] for r in rows),
                        allocs=statistics.median(r['allocs'] for r in rows), runs=rows)
            for label, rows in groups.items()}


def fmt(v, digits=2):
    return f'{v:,.{digits}f}'.replace(',', ' ').replace('.', ',')


def table(headers, rows):
    return '\n'.join(['| ' + ' | '.join(headers) + ' |', '| ' + ' | '.join(['---'] * len(headers)) + ' |'] +
                     ['| ' + ' | '.join(map(str, row)) + ' |' for row in rows])


def main():
    matrix, holdout = load_rows('matrix.jsonl'), load_rows('holdout.jsonl')
    api, parallel = parse_bench('api.txt'), parse_bench('parallel.txt')
    assert len(matrix) == 114 * 4 and len(holdout) == 7 * 4
    assert len(api) == 411 and len(parallel) == 48, (len(api), len(parallel))
    assert all(r['repeats'] == 5 for r in [*api.values(), *parallel.values()])
    cases = collections.defaultdict(dict)
    for r in matrix:
        assert sum(r['counts']) == r['samples'] and len(r['runs_ns']) == 5
        assert statistics.median(r['runs_ns']) == r['median_ns']
        cases[r['case'], r['stream']][r['implementation']] = r
    for r in holdout:
        assert sum(r['counts']) == r['samples']
    counts = {name: dict(matrix_samples=sum(r['samples'] for r in matrix if r['implementation'] == name),
                         matrix_failures=sum(not r['pass'] for r in matrix if r['implementation'] == name),
                         holdout_samples=sum(r['samples'] for r in holdout if r['implementation'] == name),
                         holdout_failures=sum(not r['pass'] for r in holdout if r['implementation'] == name)) for name in NAMES}
    wins = sum(g['vrh-ring']['median_ns'] == min(r['median_ns'] for r in g.values()) for g in cases.values())
    summary = dict(cases=len(cases), counts=counts, vrh_fastest_cases=wins, api=api, parallel=parallel)
    (OUT / 'summary.json').write_text(json.dumps(summary, indent=2) + '\n')
    with (OUT / 'owner.csv').open('w', newline='') as f:
        fields = ['case', 'stream', 'implementation', 'nodes', 'samples', 'median_ns',
                  'max_pp', 'max_z', 'chi_square', 'cv_over_noise', 'pass']
        writer = csv.DictWriter(f, fieldnames=fields, extrasaction='ignore')
        writer.writeheader()
        writer.writerows(matrix + holdout)

    def bench(key):
        return api.get(key) or api[key + '-1']

    def par(n, name, cpu):
        key = f'BenchmarkParallel/n={n}/{name}'
        return parallel.get(key + f'-{cpu}') or (parallel[key] if cpu == 1 else None)

    sizes = [1, 2, 3, 7, 15, 16, 17, 31, 32, 63, 64, 65, 256, 1024, 4096, 16384]
    selected = [3, 16, 17, 64, 256, 1024, 16384]
    owner_rows = []
    for n in sizes:
        g = cases[f'scattered-{n}', 'random']
        owner_rows.append([n] + [fmt(g[name]['median_ns']) for name in NAMES] +
                          [fmt(g[NAMES[1]]['median_ns'] / g[NAMES[0]]['median_ns']) + '×'])

    build_rows, memory_rows, direct_rows = [], [], []
    for n in selected:
        rs = [bench(f'BenchmarkBuild/n={n}/{name}') for name in NAMES]
        build_rows.append([n] + [fmt(r['ns'] / 1000, 3) for r in rs])
        memory_rows.append([n] + [f"{fmt(r['bytes'], 0)} / {fmt(r['allocs'], 0)}" for r in rs])
        d = bench(f'BenchmarkDirectNew/n={n}')
        direct_rows.append([n, fmt(d['ns']/1000, 3), fmt(d['bytes'], 0), fmt(d['allocs'], 0),
                            fmt((8 if n <= 16 else 16)*n, 0)])

    batch_rows = []
    for n in [3, 16, 17, 256, 16384]:
        for length in [16, 64, 1024]:
            batch_rows.append([n, length] + [fmt(bench(f'BenchmarkBatch/n={n}/k={length}/{name}/reuse')['ns'] / length)
                                            for name in NAMES])
    backend_rows = [[n] + [fmt(bench(f'BenchmarkBackend/n={n}/{mode}')['ns'])
                            for mode in ['auto', 'scalar', 'reference']]
                    for n in [3, 16, 17, 64, 1024, 16384]]
    rebuild_rows = [[n, q] + [fmt(bench(f'BenchmarkRebuild/n={n}/q={q}/{name}')['ns']/1000, 3)
                              for name in NAMES]
                    for n in [3, 17, 256, 16384] for q in [1, 64, 4096]]
    batch_memory_rows = [[length] + [f"{fmt(bench(f'BenchmarkBatch/n=17/k={length}/vrh-ring/{mode}')[key], 0)}"
                                    for mode in ['reuse', 'nil', 'short', 'inplace'] for key in ['bytes', 'allocs']]
                         for length in [0, 1, 16, 64, 1024]]
    parallel_rows = []
    for n in [16, 256, 16384]:
        for cpu in [1, 2, 4, 6]:
            parallel_rows.append([n, cpu] + [fmt(1000 / par(n, name, cpu)['ns'], 3) for name in NAMES])
    holdout_groups = collections.defaultdict(dict)
    for r in holdout:
        holdout_groups[r['case']][r['implementation']] = r
    quality_rows = []
    for case, g in holdout_groups.items():
        quality_rows.append([case.removeprefix('holdout-')] +
                            [f"{fmt(g[name]['max_z'])} / {'OK' if g[name]['pass'] else 'FAIL'}" for name in NAMES])
    triple_rows = [[label] + [fmt(c/r['samples']*100, 4)+'%' for c in r['counts']]
                   for name,label in zip(NAMES,SHORT)
                   for r in [holdout_groups['holdout-v2'][name]]]
    lost_cases = []
    for (case,stream),g in cases.items():
        winner=min(g,key=lambda name:g[name]['median_ns'])
        if winner != 'vrh-ring':
            lost_cases.append(f"`{case}/{stream}`: VRH {fmt(g['vrh-ring']['median_ns'])} нс, "
                              f"{winner} {fmt(g[winner]['median_ns'])} нс")

    v = counts['vrh-ring']
    appendix = [
        '## Проверка текущей библиотеки через публичный API — 13 октября 2026',
        'Этот прогон измеряет текущий `vrh.Ring` N3 через `vrh.New`, `Owner` и `OwnerBatch`. '
        'Старая таблица выше относится к исследовательскому прототипу. Проверена рабочая копия исходников; '
        'её SHA-256 сохранены вместе с результатами.',
        'AMD Ryzen 5 2600, Linux amd64, Go 1.27.1. Однопоточные замеры закреплены за CPU 4, '
        '`GOMAXPROCS=1`; частота CPU не фиксировалась. Для `Owner` — медиана пяти полных проходов, '
        'для API-бенчмарков — медиана пяти замеров длительностью не менее 100 мс. Это время операции '
        'или среднее по потоку, не p95/p99 отдельного запроса.',
        f'Матрица: **114 сценариев**, 16 размеров от 1 до 16 384 узлов, последовательные и разреженные ID, '
        f'случайные, последовательные и меняющиеся в старших 32 битах ключи, шесть известных трудных троек. '
        f'На реализацию — {fmt(v["matrix_samples"], 0)} наблюдений для распределения, те же потоки '
        f'пять раз для времени и ещё {fmt(v["holdout_samples"], 0)} наблюдений в независимой проверке.',
        '### Выбор владельца',
        'Разреженные ID, случайные `uint64` ключи; **нс/Owner, меньше лучше**. '
        'Строковые/байтовые представления ключей и кольца подготовлены заранее; хеширование запроса, '
        'поиск и возврат исходного ID включены. Последний столбец — отношение времени go-rendezvous к VRH: '
        'значение больше 1 означает преимущество VRH.',
        table(['Узлов'] + SHORT + ['go / VRH'], owner_rows),
        f'VRH имеет минимальную медиану в **{wins} из 114** сценариев полной матрицы. '
        'Это результат на этой машине и этих потоках, а не гарантия преимущества на другом CPU.',
        ('Случаи с меньшей медианой у конкурента: ' + '; '.join(lost_cases) +
         '. Разброс пяти проходов сохранён в полном отчёте; близкие медианы не устанавливают '
         'статистически значимое преимущество.') if lost_cases else
        'Разброс пяти проходов сохранён в полном отчёте; близкие медианы не устанавливают статистически значимое преимущество.',
        '¹ NSPCC scan — локальный адаптер, проверенный против первого элемента публичного `Sort`; '
        'XXH64 pair — простой эталонный скан. go-rendezvous измеряется через публичный `Lookup`. '
        'Во всех сравнительных Owner-вызовах нормализован результат к `uint64`.',
        'Отдельный benchmark путей VRH, **нс/Owner**. `OwnerScalar` — прямой эталонный Go-скан '
        'без подготовки ID; `BackendScalar` — настройка кольца, которая на amd64 использует '
        'подготовленный native-путь для больших колец.',
        table(['Узлов', 'Auto', 'Scalar', 'OwnerScalar'], backend_rows),
        'На 17 узлах в этом API-прогоне прямой `OwnerScalar` быстрее `Auto`; переход к native-пути '
        'после 16 узлов не означает, что выбранный порог оптимален для каждого размера и CPU. '
        'Это отдельный результат измерения, алгоритм по ходу прогона не перенастраивался.',
        '### Создание кольца и выделения памяти',
        'Сравнение создания адаптеров из одинакового `[]uint64`, **мкс/создание**. Включены преобразования '
        'ID и собственные проверки конструкторов. VRH проверяет дубликаты; семантика проверок конкурентов '
        'не уравнивалась. Общая обёртка измерительного стенда тоже входит в эти строки.',
        table(['Узлов'] + SHORT, build_rows),
        '**Выделенные байты / число выделений на создание** тех же адаптеров:',
        table(['Узлов'] + SHORT, memory_rows),
        'Это суммарные выделения во время создания, включая временные структуры; это не размер живого '
        'кольца после сборки мусора. Отдельно — прямой `vrh.New` без сравнительной обёртки:',
        table(['Узлов', 'New, мкс', 'Б/создание', 'Выделений', 'Данные массивов ID, Б²'], direct_rows),
        '² Полезная ёмкость массивов по коду: `8×N` до 16 узлов и `16×N` после 16. '
        'Сюда не входят структуры, заголовки slice и округление аллокатора.',
        '### Пакетная обработка',
        '**Нс/ключ**, буфер результата используется повторно. У VRH вызван `OwnerBatch`, '
        'у остальных — цикл публичного/адаптированного Owner. Это отдельный benchmark с повторяемым '
        'пакетом; его числа нельзя подменять длиннопоточным Owner выше.',
        table(['Узлов', 'Ключей в пакете'] + SHORT, batch_rows),
        'Проверены также пакеты 0 и 1, `dst=nil`, слишком короткий `dst` и запись на месте. '
        'Запись на месте проверена на правильность отдельно; повторные замеры используют предыдущие '
        'результаты как новые ключи, поэтому их скорость не включена в сравнительную таблицу.',
        'В замерах Owner зафиксировано 0 Б/op и 0 allocs/op. У VRH отсутствие выделений '
        'дополнительно проверено через `AllocsPerRun` на всех 260 размерах из теста корректности.',
        'Выделения на один `OwnerBatch` при 17 узлах; **байты и число выделений**:',
        table(['Ключей', 'Reuse, Б', 'allocs', 'Nil, Б', 'allocs', 'Short, Б', 'allocs', 'In-place, Б', 'allocs'], batch_memory_rows),
        '### Замена узла и последующие запросы',
        'Один ID заменяется, затем создаётся новое кольцо и выполняются Q запросов. '
        '**Мкс на весь цикл**; для всех реализаций измеряется полное перестроение, '
        'а не инкрементальное обновление.',
        table(['Узлов', 'Q'] + SHORT, rebuild_rows),
        '### Параллельное чтение',
        '**Миллионов Owner/с суммарно**, больше лучше. Одно неизменяемое кольцо, '
        '`GOMAXPROCS=1/2/4/6`, доступны шесть физических ядер CPU 0–5, пять замеров по ≥200 мс. '
        'Это совокупная пропускная способность, не задержка одного запроса.',
        table(['Узлов', 'GOMAXPROCS'] + SHORT, parallel_rows),
        '### Распределение и корректность',
        'Независимая проверка: по **1 048 576 новых случайных ключей** на каждый набор. '
        'В ячейке максимальный `Z` / результат двух порогов: `maxZ ≤ 7` и '
        '`χ² ≤ (N−1)+8√(2(N−1))`. На больших кольцах также важна суммарная χ², '
        'поэтому один maxZ не определяет результат. Метки v1/v2/Murmur/R2/R3 обозначают '
        'наборы узлов из прежних контрпримеров; во всех ячейках VRH исполняется нынешняя N3.',
        table(['Набор'] + SHORT, quality_rows),
        'Для прежнего контрпримера v2 `{0, 2^29, 2^59}` те же независимые наблюдения дают '
        'следующие доли владельцев (идеал — примерно 33,3333%):',
        table(['Реализация', '0', '2^29', '2^59'], triple_rows),
        'Число превышений заданных порогов по всем реализациям:',
        table(['Реализация', 'Основная матрица, из 114', 'Независимая проверка, из 7'],
              [[label, counts[name]['matrix_failures'], counts[name]['holdout_failures']]
               for name,label in zip(NAMES,SHORT)]),
        f'У VRH отказов: **{v["matrix_failures"]}/114** в основной матрице и '
        f'**{v["holdout_failures"]}/7** в независимой проверке. На большом кольце независимая '
        'выборка даёт в среднем 64 ключа на узел; малый относительный перекос она всё ещё может пропустить. '
        'Эти проверки не доказывают универсальную minwise-гарантию.',
        'В основной скоростной матрице у кольца на 16 384 узлах всего один ключ на узел в среднем. '
        'На такой разреженной гистограмме порог, выраженный через нормальное Z, не является '
        'надёжным тестом хвостовых вероятностей; его превышение само по себе не устанавливает '
        'систематический перекос. Все превышения и исходные счётчики приведены в полном отчёте.',
        'Проверены все размеры 1–257 и отдельно 1024, 4096, 16 384; совпадение `Auto`, `Scalar`, '
        '`OwnerScalar`, порядка ID и batch; изоляция входного массива и результата `Nodes()`; '
        'стабильность при добавлении/удалении узла; конкурентное чтение под `-race`. '
        '`BackendAVX2` и неизвестный backend возвращают ошибку. `Auto` и `Scalar` выбирают одну '
        'реализацию; различия их времени отражают вариацию условий замера, а не разные алгоритмы.',
        'Обычные тесты, `-race` и `go vet` прошли для библиотеки и стенда. Переносимый путь '
        'дополнительно выполнен при `GOARCH=386`; arm64 проверен сборкой, без исполнения.',
        'Полная матрица времени с разбросом, batch-режимы, смена узла с перестроением и '
        '1/64/4096 запросами, метаданные, все исходные счётчики и команды сохранены в '
        '[отчёте production-аудита](../../VRH/docs/production-audit-2026-10-13.md) и '
        '[протоколе](../../VRH/benchmarks/production/PROTOCOL.md).'
    ]
    appendix_text = '\n\n'.join(appendix) + '\n'
    (OUT / 'post-appendix.md').write_text(appendix_text)
    report = ['# Аудит текущей VRH N3 — 2026-10-13',
              'Исходники и окружение: [SHA-256](results/production-2026-10-13/sources.json), '
              '[окружение](results/production-2026-10-13/environment.txt), '
              '[зависимости](results/production-2026-10-13/modules.txt), '
              '[протокол](../benchmarks/production/PROTOCOL.md).',
              appendix_text.replace('## Проверка текущей библиотеки через публичный API — 13 октября 2026', '## Сводка')
              .replace('Старая таблица выше относится к исследовательскому прототипу.',
                       'Архивные результаты PreparedN3 относятся к исследовательскому прототипу.')
              .replace('../../VRH/docs/production-audit-2026-10-13.md', 'production-audit-2026-10-13.md')
              .replace('../../VRH/benchmarks/production/PROTOCOL.md', '../benchmarks/production/PROTOCOL.md'),
              '## Полная матрица Owner',
              'Время в нс/Owner: медиана [минимум; максимум] пяти средних по целому потоку. '
              'Для каждого случая распределение считалось отдельным проходом.']
    full = []
    for (case, stream), g in cases.items():
        full.append([case, stream, g[NAMES[0]]['samples']] +
                    [f"{fmt(g[name]['median_ns'])} [{fmt(min(g[name]['runs_ns']))}; {fmt(max(g[name]['runs_ns']))}]" for name in NAMES])
    report.append(table(['Случай', 'Ключи', 'Наблюдений'] + SHORT, full))
    report += ['## Пороги распределения по всем реализациям',
               table(['Реализация', 'Отказов основной матрицы / 114', 'Отказов holdout / 7'],
                     [[n, counts[n]['matrix_failures'], counts[n]['holdout_failures']] for n in NAMES])]
    failures = [r for r in matrix + holdout if not r['pass']]
    if failures:
        report += ['### Все превышения порогов', table(['Набор', 'Поток', 'Реализация', 'max Z', 'χ²'],
                   [[r['case'], r['stream'], r['implementation'], fmt(r['max_z'], 6), fmt(r['chi_square'])] for r in failures])]
    report += ['## Все API-бенчмарки',
               'Единица — один вызов или один полный цикл (для Batch — весь пакет, Rebuild — создание '
               'и все запросы). Параллельные ns/op — обратная совокупная пропускная способность. '
               'Сохранены все пять повторов; значения ниже — медианы.',
               table(['Benchmark', 'нс/op', 'min–max', 'Б/op', 'allocs/op'],
                     [[label, fmt(r['ns']), f"{fmt(r['min_ns'])}–{fmt(r['max_ns'])}", fmt(r['bytes'], 0), fmt(r['allocs'], 0)]
                      for label, r in [*api.items(), *parallel.items()]]),
               '## Сырые результаты',
               '\n'.join(f'- [{name}](results/production-2026-10-13/{name})' for name in
                         ['matrix.jsonl', 'holdout.jsonl', 'api.txt', 'parallel.txt', 'summary.json', 'owner.csv',
                          'production-tests.txt', 'production-race.txt', 'production-vet.txt',
                          'audit-tests.txt', 'audit-race.txt', 'audit-vet.txt', 'production-386.txt', 'arm64-build.txt'])]
    (ROOT / 'docs/production-audit-2026-10-13.md').write_text('\n\n'.join(report) + '\n')
    print(json.dumps(dict(counts=counts, fastest=wins, api_groups=len(api), parallel_groups=len(parallel)), indent=2))


if __name__ == '__main__':
    main()
