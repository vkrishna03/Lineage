"""
Lineage SDK - Epistemic transparency for AI systems.

Simple API:
    import lineage

    lineage.init(project="my-app", actor_name="my-service", actor_type="service")

    @lineage.track("recommendation", intent="suggestion")
    def recommend_price(data):
        return {"price": 26.99, "confidence": 0.85}

    lineage.emit("execution", "execution", {"action": "done"}, confidence=1.0)

Low-level API:
    from lineage import LineageClient
    client = LineageClient(base_url="http://localhost:8080")
"""

from __future__ import annotations

from typing import TYPE_CHECKING, Any

from .client import LineageClient
from .enums import ActorType, ArtifactRole, CorrectionType, Intent, ScoreCategory, ScoreType
from .exceptions import LineageError, NotFoundError, ServerError, ValidationError
from .models import Actor, Artifact, Event, EventType, HealthStatus, Lineage, Score, Scope
from ._state import _global_state
from ._decorators import track
from ._context import span

if TYPE_CHECKING:
    from .models import Event

__all__ = [
    # Simple API
    "init",
    "emit",
    "track",
    "span",
    # Low-level client
    "LineageClient",
    # Enums
    "ActorType",
    "Intent",
    "CorrectionType",
    "ScoreType",
    "ScoreCategory",
    "ArtifactRole",
    # Models
    "Scope",
    "Actor",
    "EventType",
    "Event",
    "Artifact",
    "Score",
    "Lineage",
    "HealthStatus",
    # Exceptions
    "LineageError",
    "NotFoundError",
    "ValidationError",
    "ServerError",
]

__version__ = "0.1.0"


def init(
    project: str,
    *,
    base_url: str = "http://localhost:8080",
    domain: str | None = None,
    environment: str | None = None,
    actor_name: str | None = None,
    actor_type: str = "service",
    auto_create: bool = True,
    wait_time: float = 2.0,
) -> None:
    """
    Initialize the Lineage SDK.

    This must be called before using emit(), track(), or span().

    Example:
        import lineage

        lineage.init(
            project="my-app",
            actor_name="my-service",
            actor_type="service"
        )

    Args:
        project: Project name for the scope
        base_url: Lineage API server URL
        domain: Optional domain for the scope
        environment: Optional environment (dev, staging, production)
        actor_name: Default actor name (auto-created if needed)
        actor_type: Default actor type (human, llm, agent, service, tool)
        auto_create: Auto-create scope and actor if they don't exist
        wait_time: Seconds to wait for async event processing (default 2.0)
    """
    _global_state.init(
        project=project,
        base_url=base_url,
        domain=domain,
        environment=environment,
        actor_name=actor_name,
        actor_type=actor_type,
        auto_create=auto_create,
        wait_time=wait_time,
    )


def emit(
    event_type: str,
    intent: str | Intent,
    payload: dict[str, Any],
    *,
    confidence: float | None = None,
    actor: tuple[str, str] | Actor | None = None,
    parent: Event | str | None = None,
    reason: str | None = None,
    wait: bool = True,
) -> Event | None:
    """
    Emit a Lineage event.

    Example:
        lineage.emit(
            "price_change",
            "execution",
            {"old": 29.99, "new": 27.99},
            confidence=1.0,
            actor=("service", "Pricing Engine")
        )

    Args:
        event_type: Name of the event type (auto-created if needed)
        intent: Event intent (exploration, suggestion, assertion, decision, execution)
        payload: Event payload data
        confidence: Optional confidence score (0.0-1.0)
        actor: Override actor as (type, name) tuple or Actor object
        parent: Parent event for lineage tracking
        reason: Optional reason/explanation
        wait: Wait for async processing and return the Event (default True)

    Returns:
        The created Event if wait=True, None otherwise
    """
    return _global_state.emit(
        event_type=event_type,
        intent=intent,
        payload=payload,
        confidence=confidence,
        actor=actor,
        parent=parent,
        reason=reason,
        wait=wait,
    )


# Expose global state properties for advanced use
def get_client() -> LineageClient:
    """Get the underlying LineageClient for advanced operations."""
    return _global_state.client


def get_scope() -> Scope:
    """Get the current scope."""
    return _global_state.scope


def get_last_event() -> Event | None:
    """Get the most recently emitted event."""
    return _global_state.last_event
