#!/usr/bin/env python3
"""Build 5,000 distinct single-turn Q&A examples for Metrin's live prompt shape."""
import hashlib
import itertools
import json
import random
from collections import Counter
from pathlib import Path

import generar_dataset_5000 as source

ROOT = Path(__file__).resolve().parent
OUT = ROOT / "data" / "metrin-5000-qa"
FULL = ROOT / "data" / "metrin-5000" / "conversaciones.jsonl"
SEED = 20260930
SYSTEM = source.SYSTEM
QUOTAS = source.QUOTAS

SALUDOS_EXTRA = [
    "Qué gusto saludarte", "Hola, vuelvo por aquí", "¿Podemos conversar un momento?",
    "Hola, quería hacerte una consulta", "¿Tienes un minuto?", "Te saludo, Metrín",
    "Hola, ¿puedo preguntarte algo?", "Buenas, necesito orientación", "¿Me acompañas con una duda?",
    "Hola, estoy revisando un tema", "Quería retomar la conversación", "Hola, ¿podemos revisar algo?",
    "¿Estás disponible para una consulta?", "Hola, necesito una explicación", "Vengo con otra pregunta",
    "Buenas, quiero revisar un punto", "Hola, ¿me orientas?", "¿Podemos ver una consulta?",
    "Hola, necesito aclarar algo", "Qué bueno encontrarte disponible",
]
RESP_SALUDO_EXTRA = [
    "Hola. Claro, dime qué necesitas revisar.", "Sí, te leo. ¿Cuál es tu consulta?",
    "Hola. Cuéntame qué quieres resolver y lo vemos.", "Claro, estoy aquí. ¿Qué tema revisamos?",
    "Hola. Dime qué información estás buscando.", "Con gusto. ¿Cuál es la duda que tienes?",
    "Hola. Cuéntame un poco más y te ayudo.", "Sí, podemos revisarlo. ¿Por dónde empezamos?",
    "Hola. ¿Qué necesitas saber?", "Claro, dime en qué puedo orientarte.",
]
AGRADECIMIENTOS_EXTRA = [
    "Gracias por tomarte el tiempo de revisarlo.", "Te agradezco que lo hayas aclarado.",
    "Gracias, esa respuesta me orienta.", "Muy amable, con eso avanzo.",
    "Gracias por la paciencia.", "Aprecio que hayas comprobado la fuente.",
    "Gracias, me queda más claro.", "Perfecto, agradezco la precisión.",
    "Muchas gracias por acompañarme con esto.", "Gracias por responder de forma directa.",
    "Te agradezco la explicación breve.", "Gracias, era el dato que buscaba.",
    "Bien, gracias por ayudarme a aclararlo.", "Gracias por distinguir esos puntos.",
    "Agradezco que señales lo que no está confirmado.", "Gracias, ya puedo seguir.",
    "Muy claro, te lo agradezco.", "Gracias por volver a revisarlo.",
    "Listo, gracias por la ayuda.", "Te agradezco la respuesta.",
]


def framed(phrases, wrappers):
    for phrase, wrapper in itertools.product(phrases, wrappers):
        yield wrapper.format(text=phrase)


def add_product(rows, category, users, answers):
    for user, answer in itertools.product(users, answers):
        rows.append((category, user, answer))


def build_rows(rng):
    rows = []
    add_product(rows, "saludos", source.SALUDOS + SALUDOS_EXTRA, source.RESP_SALUDO_EXTRA + RESP_SALUDO_EXTRA)

    add_product(rows, "aclaraciones", framed(source.VAGAS, [
        "{text}", "Tengo una consulta: {text}", "Necesito ayuda con esto: {text}",
        "Quiero orientarme. {text}",
    ]), source.PIDE_ACLARAR)

    add_product(rows, "correcciones", framed(source.QUEJAS, [
        "{text}", "Quiero corregir algo: {text}", "No estoy de acuerdo: {text}",
    ]), source.RESP_QUEJA)

    add_product(rows, "reformulaciones", framed(source.NO_CLARO, [
        "{text}", "Necesito una explicación. {text}", "Sobre tu respuesta: {text}",
        "Por favor, ayúdame a entenderlo. {text}",
    ]), source.PREGUNTA_REFORMULA)

    add_product(rows, "seguimiento", framed(source.SEGUIMIENTOS, [
        "{text}", "Una pregunta más: {text}", "Para continuar, {text}",
    ]), source.PREGUNTA_FOCO)

    source_answers = []
    for _, summary in source.FUENTES:
        lower = summary[0].lower() + summary[1:]
        source_answers.extend([
            summary,
            f"En resumen, {lower}",
            f"La idea central es que {lower}",
            f"Dicho de otra forma, {lower}",
        ])
    users = (
        f"La fuente dice: «{snippet}» {ask}"
        for snippet, _ in source.FUENTES
        for ask in source.PIDE_RESUMEN
    )
    add_product(rows, "texto_aportado", users, source_answers)

    add_product(rows, "limites", framed(source.TEMAS_AJENOS, [
        "{text}", "Una consulta: {text}", "Por favor, {text}",
        "Tengo otro pedido: {text}", "¿Puedes ayudarme con esto? {text}",
    ]), source.REDIRIGE)

    add_product(rows, "cierres", source.AGRADECIMIENTOS + AGRADECIMIENTOS_EXTRA, source.RESP_GRACIAS)

    step_users = []
    step_answers = []
    for i, steps in enumerate(source.PASOS):
        joined = "; ".join(steps)
        capitalized = tuple(s[0].upper() + s[1:] for s in steps)
        summary = source.RESUMENES_PASOS[i]
        for ask, response in itertools.product(source.PIDE_PASOS, source.RESP_PASOS):
            question = ask.format(joined)
            answer = response.format(*capitalized, summary)
            step_users.append(question)
            step_answers.append(answer)
    add_product(rows, "organizar_pasos", framed(step_users, [
        "{text}", "Por favor, {text}", "Necesito ordenar esto. {text}",
    ]), step_answers)

    rng.shuffle(rows)
    selected = []
    signatures = set()
    category_counts = Counter()
    for category, user, answer in rows:
        signature = hashlib.sha256((user + "\0" + answer).encode("utf-8")).hexdigest()
        if signature in signatures or category_counts[category] >= QUOTAS[category]:
            continue
        signatures.add(signature)
        category_counts[category] += 1
        selected.append({"category": category, "user": user, "assistant": answer})
    if category_counts != Counter(QUOTAS):
        raise ValueError(f"No se alcanzaron las cuotas con pares únicos: {category_counts}")
    return selected


def main():
    rng = random.Random(SEED)
    rows = build_rows(rng)
    rng.shuffle(rows)
    splits = {
        "train": rows[:4500],
        "valid": rows[4500:4750],
        "test": rows[4750:],
    }
    OUT.mkdir(parents=True, exist_ok=True)
    manifest = {
        "seed": SEED,
        "total_unique_qa": len(rows),
        "source_dialogue_count": 5000,
        "format": "single user-assistant exchange",
        "system": SYSTEM,
        "categories": dict(sorted(QUOTAS.items())),
        "splits": {},
    }
    all_signatures = set()
    for split, items in splits.items():
        path = OUT / f"{split}.jsonl"
        with path.open("w", encoding="utf-8") as f:
            for row in items:
                messages = [
                    {"role": "system", "content": SYSTEM},
                    {"role": "user", "content": row["user"]},
                    {"role": "assistant", "content": row["assistant"]},
                ]
                signature = hashlib.sha256((row["user"] + "\0" + row["assistant"]).encode("utf-8")).hexdigest()
                if signature in all_signatures:
                    raise ValueError("Repeated Q&A across splits")
                all_signatures.add(signature)
                f.write(json.dumps({"messages": messages}, ensure_ascii=False) + "\n")
        manifest["splits"][split] = {
            "examples": len(items),
            "sha256": hashlib.sha256(path.read_bytes()).hexdigest(),
            "categories": dict(sorted(Counter(row["category"] for row in items).items())),
        }
    if len(all_signatures) != 5000:
        raise ValueError(f"Esperaba 5.000 pares únicos; hay {len(all_signatures)}")
    manifest["unique_qa_pairs"] = len(all_signatures)
    if FULL.exists():
        manifest["full_dialogues_sha256"] = hashlib.sha256(FULL.read_bytes()).hexdigest()
    (OUT / "manifest.json").write_text(json.dumps(manifest, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print(json.dumps(manifest, ensure_ascii=False, indent=2))


if __name__ == "__main__":
    main()
