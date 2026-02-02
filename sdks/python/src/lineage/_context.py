"""Context manager for span-based event tracking."""

from __future__ import annotations

from typing import Any

from .enums import Intent
from .models import Event
from ._state import _global_state


class Span:
    """
    Context manager for tracking a block of code as a Lineage event.

    Example:
        with lineage.span("data_processing", intent="assertion") as span:
            result = process_data()
            span.confidence = 0.99
            span.payload = {"rows_processed": 1000}
    """

    def __init__(
        self,
        event_type: str,
        intent: str | Intent,
        *,
        actor: tuple[str, str] | None = None,
        parent: Event | None = None,
    ):
        self.event_type = event_type
        self.intent = intent
        self.actor = actor
        self.parent = parent

        # Set during span execution
        self.payload: dict[str, Any] = {}
        self.confidence: float | None = None
        self.reason: str | None = None

        # Result
        self._event: Event | None = None

    def __enter__(self) -> Span:
        return self

    def __exit__(self, exc_type: Any, exc_val: Any, exc_tb: Any) -> None:
        # Don't emit on exception
        if exc_type is not None:
            return

        # Emit the event
        self._event = _global_state.emit(
            event_type=self.event_type,
            intent=self.intent,
            payload=self.payload,
            confidence=self.confidence,
            actor=self.actor,
            parent=self.parent,
            reason=self.reason,
        )

    @property
    def event(self) -> Event | None:
        """Get the emitted event (available after exiting the context)."""
        return self._event


def span(
    event_type: str,
    intent: str | Intent,
    *,
    actor: tuple[str, str] | None = None,
    parent: Event | None = None,
) -> Span:
    """
    Create a span context manager for tracking a block of code.

    Example:
        with lineage.span("data_processing", intent="assertion") as span:
            result = process_data()
            span.confidence = 0.99
            span.payload = {"rows_processed": 1000}

    Args:
        event_type: Name of the event type (auto-created if needed)
        intent: Event intent (exploration, suggestion, assertion, decision, execution)
        actor: Optional (type, name) tuple to override default actor
        parent: Optional parent event for lineage tracking
    """
    return Span(event_type, intent, actor=actor, parent=parent)
