"""Global state management for the simple Lineage API."""

from __future__ import annotations

import time
from threading import Lock
from typing import Any
from uuid import UUID

from .client import LineageClient
from .enums import ActorType, Intent
from .models import Actor, Event, EventType, Scope


class _GlobalState:
    """Singleton holding global SDK state."""

    def __init__(self):
        self._lock = Lock()
        self._initialized = False
        self._client: LineageClient | None = None
        self._scope: Scope | None = None
        self._default_actor: Actor | None = None
        self._event_types: dict[str, EventType] = {}
        self._actors: dict[tuple[str, str], Actor] = {}  # (type, name) -> Actor
        self._last_event: Event | None = None
        self._wait_time: float = 2.0  # seconds to wait for async processing

    def init(
        self,
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
        """Initialize the SDK with project and optional default actor."""
        with self._lock:
            self._client = LineageClient(base_url=base_url)
            self._wait_time = wait_time

            if auto_create:
                # Create or find scope
                self._scope = self._client.scopes.create(
                    project=project,
                    domain=domain,
                    environment=environment,
                )

                # Create default actor if specified
                if actor_name:
                    self._default_actor = self._get_or_create_actor(actor_type, actor_name)

            self._initialized = True

    def _ensure_initialized(self) -> None:
        """Raise error if not initialized."""
        if not self._initialized:
            raise RuntimeError("lineage.init() must be called first")

    def _get_or_create_actor(
        self, actor_type: str, actor_name: str, metadata: dict | None = None
    ) -> Actor:
        """Get or create an actor by type and name."""
        key = (actor_type, actor_name)
        if key in self._actors:
            return self._actors[key]

        actor = self._client.actors.create(
            type=ActorType(actor_type),
            name=actor_name,
            metadata=metadata,
        )
        self._actors[key] = actor
        return actor

    def _get_or_create_event_type(
        self, name: str, allowed_intents: list[Intent] | None = None
    ) -> EventType:
        """Get or create an event type by name."""
        if name in self._event_types:
            return self._event_types[name]

        # Default: allow all intents
        if allowed_intents is None:
            allowed_intents = list(Intent)

        event_type = self._client.event_types.create(
            name=name,
            version="1.0",
            allowed_intents=allowed_intents,
        )
        self._event_types[name] = event_type
        return event_type

    def _resolve_actor(
        self, actor: tuple[str, str] | Actor | None
    ) -> Actor:
        """Resolve actor from tuple (type, name), Actor object, or use default."""
        if actor is None:
            if self._default_actor is None:
                raise ValueError("No actor specified and no default actor set")
            return self._default_actor

        if isinstance(actor, Actor):
            return actor

        if isinstance(actor, tuple) and len(actor) == 2:
            return self._get_or_create_actor(actor[0], actor[1])

        raise ValueError(f"Invalid actor: {actor}")

    def emit(
        self,
        event_type: str,
        intent: str | Intent,
        payload: dict[str, Any],
        *,
        confidence: float | None = None,
        actor: tuple[str, str] | Actor | None = None,
        parent: Event | UUID | str | None = None,
        reason: str | None = None,
        wait: bool = True,
    ) -> Event | None:
        """Emit an event."""
        self._ensure_initialized()

        # Resolve intent
        if isinstance(intent, str):
            intent = Intent(intent)

        # Resolve actor
        resolved_actor = self._resolve_actor(actor)

        # Get or create event type
        et = self._get_or_create_event_type(event_type, [intent])

        # Resolve parent
        parent_ids = None
        if parent:
            if isinstance(parent, Event):
                parent_ids = [parent.id]
            elif isinstance(parent, (UUID, str)):
                parent_ids = [str(parent)]

        # Extract confidence from payload if present
        if confidence is None and isinstance(payload, dict):
            confidence = payload.pop("confidence", None)

        # Create event
        self._client.events.create(
            scope_id=self._scope.id,
            actor_id=resolved_actor.id,
            event_type_id=et.id,
            intent=intent,
            payload=payload,
            confidence=confidence,
            parent_event_ids=parent_ids,
            reason=reason,
        )

        # Wait for async processing and fetch event
        if wait and self._wait_time > 0:
            time.sleep(self._wait_time)
            events = self._client.events.list(scope_id=self._scope.id)
            for event in events:
                if (
                    event.intent == intent
                    and event.actor_id == resolved_actor.id
                ):
                    self._last_event = event
                    return event

        return None

    @property
    def client(self) -> LineageClient:
        """Get the underlying client for advanced operations."""
        self._ensure_initialized()
        return self._client

    @property
    def scope(self) -> Scope:
        """Get the current scope."""
        self._ensure_initialized()
        return self._scope

    @property
    def last_event(self) -> Event | None:
        """Get the last emitted event."""
        return self._last_event


# Global singleton
_global_state = _GlobalState()
