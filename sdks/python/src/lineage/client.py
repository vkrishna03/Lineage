"""Lineage API client."""

from __future__ import annotations

from typing import Any
from uuid import UUID

import httpx

from .enums import ActorType, ArtifactRole, CorrectionType, Intent, ScoreType
from .exceptions import LineageError, NotFoundError, ServerError, ValidationError
from .models import Actor, Artifact, Event, EventType, HealthStatus, Lineage, Score, Scope


def _handle_response(response: httpx.Response) -> dict[str, Any]:
    """Handle API response and raise appropriate exceptions."""
    if response.status_code >= 500:
        raise ServerError(response.text)
    if response.status_code == 404:
        data = response.json()
        raise NotFoundError("Resource", data.get("error", "unknown"))
    if response.status_code == 400:
        data = response.json()
        raise ValidationError(data.get("error", "Validation failed"))
    if response.status_code >= 400:
        raise LineageError(response.text, response.status_code)
    return response.json()


class ScopesResource:
    """Scopes API resource."""

    def __init__(self, http: httpx.Client):
        self._http = http

    def create(
        self,
        project: str,
        domain: str | None = None,
        environment: str | None = None,
    ) -> Scope:
        """Create a new scope."""
        payload = {"project": project}
        if domain:
            payload["domain"] = domain
        if environment:
            payload["environment"] = environment

        response = self._http.post("/api/v1/scopes", json=payload)
        data = _handle_response(response)
        return Scope.model_validate(data)

    def get(self, id: UUID | str) -> Scope:
        """Get a scope by ID."""
        response = self._http.get(f"/api/v1/scopes/{id}")
        data = _handle_response(response)
        return Scope.model_validate(data)

    def list(self) -> list[Scope]:
        """List all scopes."""
        response = self._http.get("/api/v1/scopes")
        data = _handle_response(response)
        return [Scope.model_validate(s) for s in data.get("scopes", [])]


class ActorsResource:
    """Actors API resource."""

    def __init__(self, http: httpx.Client):
        self._http = http

    def create(
        self,
        type: ActorType,
        name: str | None = None,
        external_id: str | None = None,
        metadata: dict[str, Any] | None = None,
    ) -> Actor:
        """Create a new actor."""
        payload: dict[str, Any] = {"type": type.value}
        if name:
            payload["name"] = name
        if external_id:
            payload["external_id"] = external_id
        if metadata:
            payload["metadata"] = metadata

        response = self._http.post("/api/v1/actors", json=payload)
        data = _handle_response(response)
        return Actor.model_validate(data)

    def get(self, id: UUID | str) -> Actor:
        """Get an actor by ID."""
        response = self._http.get(f"/api/v1/actors/{id}")
        data = _handle_response(response)
        return Actor.model_validate(data)

    def list(self) -> list[Actor]:
        """List all actors."""
        response = self._http.get("/api/v1/actors")
        data = _handle_response(response)
        return [Actor.model_validate(a) for a in data.get("actors", [])]


class EventTypesResource:
    """Event Types API resource."""

    def __init__(self, http: httpx.Client):
        self._http = http

    def create(
        self,
        name: str,
        version: str,
        description: str | None = None,
        payload_schema: dict[str, Any] | None = None,
        allowed_intents: list[Intent] | None = None,
    ) -> EventType:
        """Create a new event type."""
        payload: dict[str, Any] = {"name": name, "version": version}
        if description:
            payload["description"] = description
        if payload_schema:
            payload["payload_schema"] = payload_schema
        if allowed_intents:
            payload["allowed_intents"] = [i.value for i in allowed_intents]

        response = self._http.post("/api/v1/event-types", json=payload)
        data = _handle_response(response)
        return EventType.model_validate(data)

    def get(self, id: UUID | str) -> EventType:
        """Get an event type by ID."""
        response = self._http.get(f"/api/v1/event-types/{id}")
        data = _handle_response(response)
        return EventType.model_validate(data)

    def list(self) -> list[EventType]:
        """List all active event types."""
        response = self._http.get("/api/v1/event-types")
        data = _handle_response(response)
        return [EventType.model_validate(et) for et in data.get("event_types", [])]


class EventsResource:
    """Events API resource."""

    def __init__(self, http: httpx.Client):
        self._http = http

    def create(
        self,
        scope_id: UUID | str,
        actor_id: UUID | str,
        event_type_id: UUID | str,
        intent: Intent,
        payload: dict[str, Any],
        reason: str | None = None,
        correction_type: CorrectionType | None = None,
        corrects_event_id: UUID | str | None = None,
        observed_at: str | None = None,
        decided_at: str | None = None,
        parent_event_ids: list[UUID | str] | None = None,
        confidence: float | None = None,
        input_artifact_ids: list[UUID | str] | None = None,
        output_artifact_ids: list[UUID | str] | None = None,
    ) -> dict[str, str]:
        """
        Create a new event.

        Events are processed asynchronously via Kafka.
        Returns a status dict with 'status' and 'message' fields.
        """
        data: dict[str, Any] = {
            "scope_id": str(scope_id),
            "actor_id": str(actor_id),
            "event_type_id": str(event_type_id),
            "intent": intent.value,
            "payload": payload,
        }
        if reason:
            data["reason"] = reason
        if correction_type:
            data["correction_type"] = correction_type.value
        if corrects_event_id:
            data["corrects_event_id"] = str(corrects_event_id)
        if observed_at:
            data["observed_at"] = observed_at
        if decided_at:
            data["decided_at"] = decided_at
        if parent_event_ids:
            data["parent_event_ids"] = [str(p) for p in parent_event_ids]
        if confidence is not None:
            data["confidence"] = confidence
        if input_artifact_ids:
            data["input_artifact_ids"] = [str(a) for a in input_artifact_ids]
        if output_artifact_ids:
            data["output_artifact_ids"] = [str(a) for a in output_artifact_ids]

        response = self._http.post("/api/v1/events", json=data)
        return _handle_response(response)

    def get(self, id: UUID | str) -> Event:
        """Get an event by ID."""
        response = self._http.get(f"/api/v1/events/{id}")
        data = _handle_response(response)
        return Event.model_validate(data)

    def list(self, scope_id: UUID | str) -> list[Event]:
        """List events by scope."""
        response = self._http.get(f"/api/v1/events?scope_id={scope_id}")
        data = _handle_response(response)
        return [Event.model_validate(e) for e in data.get("events", [])]

    def get_lineage(self, id: UUID | str) -> Lineage:
        """Get lineage (parents and children) for an event."""
        response = self._http.get(f"/api/v1/events/{id}/lineage")
        data = _handle_response(response)
        return Lineage.model_validate(data)


class ArtifactsResource:
    """Artifacts API resource."""

    def __init__(self, http: httpx.Client):
        self._http = http

    def create(
        self,
        scope_id: UUID | str,
        content_hash: str,
        content_type: str,
        uri: str | None = None,
        metadata: dict[str, Any] | None = None,
    ) -> Artifact:
        """Create a new artifact."""
        payload: dict[str, Any] = {
            "scope_id": str(scope_id),
            "content_hash": content_hash,
            "content_type": content_type,
        }
        if uri:
            payload["uri"] = uri
        if metadata:
            payload["metadata"] = metadata

        response = self._http.post("/api/v1/artifacts", json=payload)
        data = _handle_response(response)
        return Artifact.model_validate(data)

    def get(self, id: UUID | str) -> Artifact:
        """Get an artifact by ID."""
        response = self._http.get(f"/api/v1/artifacts/{id}")
        data = _handle_response(response)
        return Artifact.model_validate(data)

    def get_by_hash(self, scope_id: UUID | str, content_hash: str) -> Artifact:
        """Get an artifact by scope and content hash (for deduplication)."""
        response = self._http.get(
            f"/api/v1/artifacts?scope_id={scope_id}&content_hash={content_hash}"
        )
        data = _handle_response(response)
        return Artifact.model_validate(data)

    def list(self, scope_id: UUID | str) -> list[Artifact]:
        """List artifacts by scope."""
        response = self._http.get(f"/api/v1/artifacts?scope_id={scope_id}")
        data = _handle_response(response)
        return [Artifact.model_validate(a) for a in data.get("artifacts", [])]

    def link_to_event(
        self,
        event_id: UUID | str,
        artifact_id: UUID | str,
        role: ArtifactRole,
    ) -> dict[str, str]:
        """Link an artifact to an event."""
        payload = {"artifact_id": str(artifact_id), "role": role.value}
        response = self._http.post(f"/api/v1/events/{event_id}/artifacts", json=payload)
        return _handle_response(response)

    def get_for_event(self, event_id: UUID | str) -> list[Artifact]:
        """Get artifacts linked to an event."""
        response = self._http.get(f"/api/v1/events/{event_id}/artifacts")
        data = _handle_response(response)
        return [Artifact.model_validate(a) for a in data.get("artifacts", [])]


class ScoresResource:
    """Scores API resource."""

    def __init__(self, http: httpx.Client):
        self._http = http

    def create(
        self,
        event_id: UUID | str,
        type: ScoreType,
        value: float,
        scored_by: UUID | str | None = None,
        reason: str | None = None,
        metadata: dict[str, Any] | None = None,
    ) -> Score:
        """Add a score to an event."""
        payload: dict[str, Any] = {"type": type.value, "value": value}
        if scored_by:
            payload["scored_by"] = str(scored_by)
        if reason:
            payload["reason"] = reason
        if metadata:
            payload["metadata"] = metadata

        response = self._http.post(f"/api/v1/events/{event_id}/scores", json=payload)
        data = _handle_response(response)
        return Score.model_validate(data)

    def list(
        self,
        event_id: UUID | str,
        type: ScoreType | None = None,
    ) -> list[Score]:
        """Get scores for an event, optionally filtered by type."""
        url = f"/api/v1/events/{event_id}/scores"
        if type:
            url += f"?type={type.value}"
        response = self._http.get(url)
        data = _handle_response(response)
        return [Score.model_validate(s) for s in data.get("scores", [])]


class LineageClient:
    """
    Lineage API client.

    Example:
        >>> client = LineageClient(base_url="http://localhost:8080")
        >>> scope = client.scopes.create(project="my-project")
        >>> actor = client.actors.create(type=ActorType.HUMAN, name="Alice")
    """

    def __init__(self, base_url: str = "http://localhost:8080", timeout: float = 30.0):
        """
        Initialize the Lineage client.

        Args:
            base_url: The base URL of the Lineage API server.
            timeout: Request timeout in seconds.
        """
        self._http = httpx.Client(base_url=base_url, timeout=timeout)
        self.scopes = ScopesResource(self._http)
        self.actors = ActorsResource(self._http)
        self.event_types = EventTypesResource(self._http)
        self.events = EventsResource(self._http)
        self.artifacts = ArtifactsResource(self._http)
        self.scores = ScoresResource(self._http)

    def health(self) -> HealthStatus:
        """Check API health status."""
        response = self._http.get("/health")
        data = _handle_response(response)
        return HealthStatus.model_validate(data)

    def close(self) -> None:
        """Close the HTTP client."""
        self._http.close()

    def __enter__(self) -> "LineageClient":
        return self

    def __exit__(self, *args: Any) -> None:
        self.close()
