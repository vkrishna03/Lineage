"""Pydantic models for the Lineage SDK."""

from __future__ import annotations

import base64
import json
from datetime import datetime
from typing import Any
from uuid import UUID

from pydantic import BaseModel, ConfigDict, field_validator

from .enums import ActorType, CorrectionType, Intent, ScoreCategory, ScoreType


def _decode_metadata(v: Any) -> dict[str, Any] | None:
    """Decode metadata that may be base64-encoded JSON (from Go API []byte)."""
    if v is None:
        return None
    if isinstance(v, dict):
        return v
    if isinstance(v, str):
        # Try to decode as base64
        try:
            decoded = base64.b64decode(v)
            return json.loads(decoded)
        except Exception:
            # Maybe it's just a JSON string
            try:
                return json.loads(v)
            except Exception:
                return None
    return None


class Scope(BaseModel):
    """A scope defines the boundary for an event chain."""

    model_config = ConfigDict(from_attributes=True)

    id: UUID
    project: str
    domain: str | None = None
    environment: str | None = None
    created_at: datetime


class Actor(BaseModel):
    """An entity that generates events."""

    model_config = ConfigDict(from_attributes=True)

    id: UUID
    type: ActorType
    external_id: str | None = None
    name: str | None = None
    metadata: dict[str, Any] | None = None
    registered_at: datetime

    @field_validator("metadata", mode="before")
    @classmethod
    def decode_metadata(cls, v: Any) -> dict[str, Any] | None:
        return _decode_metadata(v)


class EventType(BaseModel):
    """Schema and configuration for a type of event."""

    model_config = ConfigDict(from_attributes=True)

    id: UUID
    name: str
    version: str
    description: str | None = None
    payload_schema: dict[str, Any] | None = None
    allowed_intents: list[str]
    is_active: bool
    created_at: datetime

    @field_validator("payload_schema", mode="before")
    @classmethod
    def decode_payload_schema(cls, v: Any) -> dict[str, Any] | None:
        return _decode_metadata(v)


class Event(BaseModel):
    """An immutable record of an action with epistemic context."""

    model_config = ConfigDict(from_attributes=True)

    id: UUID
    scope_id: UUID
    actor_id: UUID
    event_type_id: UUID
    scope_sequence: int
    intent: Intent
    reason: str | None = None
    correction_type: CorrectionType | None = None
    corrects_event_id: UUID | None = None
    observed_at: datetime | None = None
    decided_at: datetime | None = None
    ingested_at: datetime
    prev_event_hash: str | None = None
    event_hash: str
    payload: dict[str, Any]

    @field_validator("payload", mode="before")
    @classmethod
    def decode_payload(cls, v: Any) -> dict[str, Any]:
        result = _decode_metadata(v)
        return result if result is not None else {}

    @field_validator("correction_type", mode="before")
    @classmethod
    def parse_correction_type(cls, v: Any) -> CorrectionType | None:
        """Handle the nested correction_type object from Go API."""
        if v is None:
            return None
        if isinstance(v, str):
            return CorrectionType(v) if v else None
        if isinstance(v, dict):
            # Go returns {"correction_type": "...", "valid": true/false}
            if not v.get("valid", False):
                return None
            ct = v.get("correction_type")
            return CorrectionType(ct) if ct else None
        return None

    @field_validator("corrects_event_id", mode="before")
    @classmethod
    def parse_corrects_event_id(cls, v: Any) -> UUID | None:
        """Handle the nested UUID object from Go API."""
        if v is None:
            return None
        if isinstance(v, UUID):
            return v
        if isinstance(v, str):
            return UUID(v) if v else None
        if isinstance(v, dict):
            # Go returns {"UUID": "...", "Valid": true/false} for pgtype.UUID
            if not v.get("Valid", False):
                return None
            uuid_str = v.get("UUID")
            return UUID(uuid_str) if uuid_str else None
        return None

    @field_validator("observed_at", "decided_at", mode="before")
    @classmethod
    def parse_nullable_timestamp(cls, v: Any) -> datetime | None:
        """Handle nullable timestamps from Go API (pgtype.Timestamptz)."""
        if v is None:
            return None
        if isinstance(v, datetime):
            return v
        if isinstance(v, str):
            return datetime.fromisoformat(v.replace("Z", "+00:00")) if v else None
        if isinstance(v, dict):
            # Go returns {"Time": "...", "Valid": true/false}
            if not v.get("Valid", False):
                return None
            time_str = v.get("Time")
            if time_str:
                return datetime.fromisoformat(time_str.replace("Z", "+00:00"))
            return None
        return None


class Artifact(BaseModel):
    """Content-addressed data consumed or produced by events."""

    model_config = ConfigDict(from_attributes=True)

    id: UUID
    scope_id: UUID
    content_hash: str
    content_type: str
    uri: str | None = None
    metadata: dict[str, Any] | None = None
    created_at: datetime

    @field_validator("metadata", mode="before")
    @classmethod
    def decode_metadata(cls, v: Any) -> dict[str, Any] | None:
        return _decode_metadata(v)


class Score(BaseModel):
    """A numeric assessment attached to an event."""

    model_config = ConfigDict(from_attributes=True)

    id: UUID
    event_id: UUID
    type: ScoreType
    value: float
    category: ScoreCategory
    scored_by: UUID | None = None
    reason: str | None = None
    metadata: dict[str, Any] | None = None
    created_at: datetime

    @field_validator("metadata", mode="before")
    @classmethod
    def decode_metadata(cls, v: Any) -> dict[str, Any] | None:
        return _decode_metadata(v)

    @field_validator("value", mode="before")
    @classmethod
    def parse_numeric_value(cls, v: Any) -> float:
        """Handle pgtype.Numeric from Go API."""
        if isinstance(v, (int, float)):
            return float(v)
        if isinstance(v, str):
            return float(v)
        if isinstance(v, dict):
            # pgtype.Numeric returns {"Int": ..., "Exp": ..., "NaN": ..., "Valid": ...}
            if v.get("Valid", False):
                int_val = v.get("Int")
                exp = v.get("Exp", 0)
                if int_val is not None:
                    return float(int_val) * (10 ** exp)
            return 0.0
        return float(v) if v else 0.0

    @field_validator("scored_by", mode="before")
    @classmethod
    def parse_scored_by(cls, v: Any) -> UUID | None:
        """Handle nullable UUID from Go API."""
        if v is None:
            return None
        if isinstance(v, UUID):
            return v
        if isinstance(v, str):
            return UUID(v) if v else None
        if isinstance(v, dict):
            if not v.get("Valid", False):
                return None
            uuid_str = v.get("UUID")
            return UUID(uuid_str) if uuid_str else None
        return None


class Lineage(BaseModel):
    """Lineage information for an event."""

    model_config = ConfigDict(from_attributes=True)

    event_id: UUID
    parents: list[Event]
    children: list[Event]


class HealthStatus(BaseModel):
    """Health check response."""

    status: str
    services: dict[str, str]
