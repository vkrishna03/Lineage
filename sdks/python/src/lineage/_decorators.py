"""Decorators for simple event tracking."""

from __future__ import annotations

import functools
from typing import Any, Callable, TypeVar

from .enums import Intent
from ._state import _global_state

F = TypeVar("F", bound=Callable[..., Any])


def track(
    event_type: str,
    *,
    intent: str | Intent,
    actor: tuple[str, str] | None = None,
    confidence: float | None = None,
) -> Callable[[F], F]:
    """
    Decorator to track function execution as a Lineage event.

    The function's return value becomes the event payload.
    If the return value is a dict with a 'confidence' key, it will be
    extracted and used as the event confidence.

    Example:
        @lineage.track("recommendation", intent="suggestion", actor=("llm", "Pricing AI"))
        def recommend_price(data):
            return {"price": 26.99, "confidence": 0.85}

    Args:
        event_type: Name of the event type (auto-created if needed)
        intent: Event intent (exploration, suggestion, assertion, decision, execution)
        actor: Optional (type, name) tuple to override default actor
        confidence: Optional fixed confidence (overridden by return value if present)
    """

    def decorator(func: F) -> F:
        @functools.wraps(func)
        def wrapper(*args: Any, **kwargs: Any) -> Any:
            # Call the function
            result = func(*args, **kwargs)

            # Prepare payload
            if isinstance(result, dict):
                payload = result.copy()
            else:
                payload = {"result": result}

            # Extract confidence from payload if present
            event_confidence = confidence
            if "confidence" in payload:
                event_confidence = payload.pop("confidence")

            # Emit event
            _global_state.emit(
                event_type=event_type,
                intent=intent,
                payload=payload,
                confidence=event_confidence,
                actor=actor,
            )

            return result

        return wrapper  # type: ignore

    return decorator
