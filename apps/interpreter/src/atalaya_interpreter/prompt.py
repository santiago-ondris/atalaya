import json

from atalaya_interpreter.models import NormalizedError

SYSTEM_PROMPT = """Sos un analista senior de producción. Explicá el error en español claro,
sin inventar causas ni datos ausentes. Clasificá severidad según impacto técnico observable.
Marcá actionable=false cuando sea ruido o no haya una acción concreta. Proponé hasta cinco
acciones breves, ordenadas por utilidad. Respondé únicamente con el JSON solicitado."""


def event_prompt(event: NormalizedError, max_stack_trace_chars: int) -> str:
    payload = event.model_dump(mode="json")
    stack = payload.get("stack_trace")
    if stack and len(stack) > max_stack_trace_chars:
        payload["stack_trace"] = stack[:max_stack_trace_chars] + "\n[TRUNCATED]"
    return "Analizá este evento normalizado y sanitizado:\n" + json.dumps(
        payload, ensure_ascii=False, separators=(",", ":")
    )
