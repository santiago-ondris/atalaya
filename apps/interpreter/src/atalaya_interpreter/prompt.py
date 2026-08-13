import json
import re

from atalaya_interpreter.models import NormalizedError

SYSTEM_PROMPT = """Sos un analista senior de producción. Explicá el error en español claro,
sin inventar causas ni datos ausentes. Clasificá severidad según impacto técnico observable.
Marcá actionable=false cuando sea ruido o no haya una acción concreta. Proponé hasta cinco
acciones breves, ordenadas por utilidad. Respondé únicamente con el JSON solicitado."""

SENSITIVE_PATTERNS = [
    (re.compile(r"(?i)(bearer\s+)[a-zA-Z0-9_\-\.=]+"), r"\1[REDACTED]"),
    (
        re.compile(
            r"(?i)(password|passwd|secret|api_key|token|auth_token)\s*[:=]\s*[\"']?[^\s\"'&,]+[\"']?"
        ),
        r"\1=[REDACTED]",
    ),
    (
        re.compile(r"(postgres|postgresql|mysql|mongodb|redis)://[^\s@]+@"),
        r"\1://[REDACTED]@",
    ),
]


def sanitize_text(text: str) -> str:
    if not text:
        return text
    for pattern, replacement in SENSITIVE_PATTERNS:
        text = pattern.sub(replacement, text)
    return text


def event_prompt(event: NormalizedError, max_stack_trace_chars: int) -> str:
    payload = event.model_dump(mode="json")
    stack = payload.get("stack_trace")
    if stack:
        stack = sanitize_text(stack)
        if len(stack) > max_stack_trace_chars:
            payload["stack_trace"] = stack[:max_stack_trace_chars] + "\n[TRUNCATED]"
        else:
            payload["stack_trace"] = stack
    if payload.get("message"):
        payload["message"] = sanitize_text(payload["message"])
    return "Analizá este evento normalizado y sanitizado:\n" + json.dumps(
        payload, ensure_ascii=False, separators=(",", ":")
    )
